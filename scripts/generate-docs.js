#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

const SWAGGER_FILE = '../src/docs/swagger.json';
const DOCS_DIR = '../docs/docs/devops/restapi/reference';

function main() {
    console.log('Generating Jekyll API documentation from swagger.json...');
    
    // Read swagger file
    const swaggerPath = path.join(__dirname, SWAGGER_FILE);
    if (!fs.existsSync(swaggerPath)) {
        console.error('Error: swagger.json not found. Run "make generate-swagger" first.');
        process.exit(1);
    }
    
    const swagger = JSON.parse(fs.readFileSync(swaggerPath, 'utf8'));
    console.log(`Found ${Object.keys(swagger.paths).length} endpoints in swagger.json`);
    
    // Group endpoints by tag
    const categories = {};
    
    Object.entries(swagger.paths).forEach(([path, methods]) => {
        Object.entries(methods).forEach(([method, details]) => {
            const tag = details.tags?.[0] || 'Other';
            
            if (!categories[tag]) {
                categories[tag] = {
                    name: tag,
                    path: tag.toLowerCase().replace(/\s+/g, '_'),
                    endpoints: []
                };
            }
            
            const endpoint = createDetailedEndpoint(path, method, details, swagger);
            categories[tag].endpoints.push(endpoint);
        });
    });
    
    // Generate main index file
    generateMainIndex(categories);
    
    // Generate individual category files
    Object.values(categories).forEach(category => {
        generateCategoryFile(category, categories);
    });
    
    console.log(`✓ Generated documentation for ${Object.keys(categories).length} categories`);
}

function createDetailedEndpoint(path, method, details, swagger) {
    const endpoint = {
        path: path,
        method: method,
        title: details.summary || `${method.toUpperCase()} ${path}`,
        description: details.description || details.summary || `${method.toUpperCase()} ${path}`,
        requires_authorization: details.security && details.security.length > 0
    };
    
    // Extract roles and claims from security
    if (details.security) {
        const roles = [];
        const claims = [];
        details.security.forEach(security => {
            Object.keys(security).forEach(key => {
                if (key.toLowerCase().includes('role')) roles.push(...security[key]);
                if (key.toLowerCase().includes('claim')) claims.push(...security[key]);
            });
        });
        if (roles.length > 0) endpoint.default_required_roles = roles;
        if (claims.length > 0) endpoint.default_required_claims = claims;
    }
    
    // Generate request examples
    endpoint.example_blocks = generateRequestExamples(path, method, details);
    
    // Generate response examples
    endpoint.response_blocks = generateResponseExamples(path, method, details, swagger);
    
    return endpoint;
}

function generateRequestExamples(path, method, details) {
    const examples = [];
    const hasBody = method.toLowerCase() === 'post' || method.toLowerCase() === 'put' || method.toLowerCase() === 'patch';
    
    // Generate cURL example (this will be the default)
    let curlExample = `curl --location '{{host}}${path.replace(/{([^}]+)}/g, '{$1}')}'`;
    
    if (details.security) {
        curlExample += ` \\\n--header 'Authorization: ******'`;
    }
    
    if (hasBody) {
        curlExample += ` \\\n--header 'Content-Type: application/json'`;
        curlExample += ` \\\n--data '${generateSampleRequestBody(details)}'`;
    }
    
    examples.push({
        title: 'cURL',
        language: 'powershell',  // Changed to powershell to match template expectations
        code_block: curlExample
    });
    
    // Generate C# example
    const csharpExample = `var client = new HttpClient();
var request = new HttpRequestMessage(HttpMethod.${method.charAt(0).toUpperCase() + method.slice(1).toLowerCase()}, "{{host}}${path}");
request.Headers.Add("Authorization", "******");
${hasBody ? `request.Content = new StringContent("${generateSampleRequestBody(details)}", Encoding.UTF8, "application/json");` : ''}
var response = await client.SendAsync(request);`;

    examples.push({
        title: 'C#',
        language: 'csharp',
        code_block: csharpExample
    });
    
    // Generate Go example
    const goExample = `package main

import (
    "fmt"
    "strings"
    "net/http"
    "io/ioutil"
)

func main() {
    url := "{{host}}${path}"
    method := "${method.toUpperCase()}"
    
    ${hasBody ? `payload := strings.NewReader("${generateSampleRequestBody(details)}")` : `payload := strings.NewReader("")`}
    
    client := &http.Client {}
    req, err := http.NewRequest(method, url, payload)
    
    if err != nil {
        fmt.Println(err)
        return
    }
    req.Header.Add("Authorization", "******")
    ${hasBody ? `req.Header.Add("Content-Type", "application/json")` : ''}
    
    res, err := client.Do(req)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer res.Body.Close()
    
    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(string(body))
}`;

    examples.push({
        title: 'Go',
        language: 'go',
        code_block: goExample
    });
    
    return examples;
}

function generateResponseExamples(path, method, details, swagger) {
    const responses = [];
    
    if (details.responses) {
        Object.entries(details.responses).forEach(([code, response]) => {
            const codeInt = parseInt(code);
            let codeDescription = 'Unknown';
            
            if (codeInt >= 200 && codeInt < 300) codeDescription = response.description || 'OK';
            else if (codeInt === 400) codeDescription = 'Bad Request';
            else if (codeInt === 401) codeDescription = 'Unauthorized'; 
            else if (codeInt === 402) codeDescription = 'Payment Required';
            else if (codeInt === 404) codeDescription = 'Not Found';
            else if (codeInt >= 500) codeDescription = 'Internal Server Error';
            else codeDescription = response.description || 'Unknown';
            
            // Generate sample response body
            let responseBody = generateSampleResponseBody(response, swagger);
            
            responses.push({
                code: code,
                code_description: codeDescription,
                code_block: responseBody
            });
        });
    }
    
    return responses;
}

function generateSampleRequestBody(details) {
    if (!details.parameters) return '{}';
    
    const bodyParam = details.parameters.find(p => p.in === 'body');
    if (!bodyParam) return '{}';
    
    // Simple body example
    return JSON.stringify({
        "key": "SomeKey",
        "secret": "SomeLongSecret"
    }, null, 2).replace(/\n/g, '\\n').replace(/"/g, '\\"');
}

function generateSampleResponseBody(response, swagger) {
    if (response.schema) {
        if (response.schema.type === 'object') {
            return `{
  "message": "Success",
  "data": {
    "id": "string",
    "name": "string",
    "status": "active"
  }
}`;
        }
        if (response.schema.type === 'array') {
            return `[
  {
    "id": "string",
    "name": "string"
  }
]`;
        }
    }
    
    // Default responses based on status code
    if (response.description && response.description.toLowerCase().includes('unauthorized')) {
        return `{
  "code": "int",
  "message": "string",
  "stack": [
    {
      "function": "string",
      "file": "string",
      "line": "int"
    }
  ]
}`;
    }
    
    return `{
  "message": "string",
  "timestamp": "2024-01-01T00:00:00Z"
}`;
}

function generateMainIndex(categories) {
    const docsDir = path.join(__dirname, DOCS_DIR);
    if (!fs.existsSync(docsDir)) {
        fs.mkdirSync(docsDir, { recursive: true });
    }
    
    const frontMatter = `---
layout: api
title: API Reference
default_host: http://localhost
api_prefix: /api
is_category_document: true
categories:
${Object.values(categories).map(cat => `    - name: ${cat.name}
      path: ${cat.path}
      endpoints:
${cat.endpoints.map(ep => `        - anchor: ${ep.path.replace(/[{}]/g, '').replace(/\//g, '_')}_${ep.method}
          method: ${ep.method}
          path: ${ep.path}
          description: ${ep.description}
          title: ${ep.title}`).join('\n')}`).join('\n')}
---

# API Reference

This page contains all API endpoints organized by category.
`;
    
    fs.writeFileSync(path.join(docsDir, 'index.md'), frontMatter);
    console.log('✓ Generated main index.md');
}

function generateCategoryFile(category, allCategories) {
    const docsDir = path.join(__dirname, DOCS_DIR);
    const categoryDir = path.join(docsDir, category.path);
    
    if (!fs.existsSync(categoryDir)) {
        fs.mkdirSync(categoryDir, { recursive: true });
    }
    
    const frontMatter = `---
layout: api
title: ${category.name}
default_host: http://localhost
api_prefix: /api
categories:
${Object.values(allCategories).map(cat => `    - name: ${cat.name}
      path: ${cat.path}
      endpoints:
${cat.endpoints.map(ep => `        - anchor: ${ep.path.replace(/[{}]/g, '').replace(/\//g, '_')}_${ep.method}
          method: ${ep.method}
          path: ${ep.path}
          description: ${ep.description}
          title: ${ep.title}`).join('\n')}`).join('\n')}
endpoints:
${category.endpoints.map(ep => generateEndpointYAML(ep)).join('\n')}
---

# ${category.name} API

This section contains all ${category.name.toLowerCase()} related endpoints.
`;
    
    fs.writeFileSync(path.join(categoryDir, 'index.md'), frontMatter);
    console.log(`✓ Generated ${category.path}/index.md`);
}

function generateEndpointYAML(endpoint) {
    const yaml = `  - path: ${endpoint.path}
    method: ${endpoint.method}
    title: ${endpoint.title}
    description: ${endpoint.description}`;
    
    let additionalFields = '';
    
    if (endpoint.requires_authorization) {
        additionalFields += '\n    requires_authorization: true';
    }
    
    if (endpoint.default_required_roles) {
        additionalFields += `\n    default_required_roles:\n${endpoint.default_required_roles.map(role => `      - ${role}`).join('\n')}`;
    }
    
    if (endpoint.default_required_claims) {
        additionalFields += `\n    default_required_claims:\n${endpoint.default_required_claims.map(claim => `      - ${claim}`).join('\n')}`;
    }
    
    if (endpoint.example_blocks && endpoint.example_blocks.length > 0) {
        additionalFields += '\n    example_blocks:';
        endpoint.example_blocks.forEach(example => {
            additionalFields += `\n      - title: ${example.title}`;
            additionalFields += `\n        language: ${example.language}`;
            additionalFields += `\n        code_block: |`;
            example.code_block.split('\n').forEach(line => {
                additionalFields += `\n          ${line}`;
            });
        });
    }
    
    if (endpoint.response_blocks && endpoint.response_blocks.length > 0) {
        additionalFields += '\n    response_blocks:';
        endpoint.response_blocks.forEach(response => {
            additionalFields += `\n      - code: ${response.code}`;
            additionalFields += `\n        code_description: ${response.code_description}`;
            additionalFields += `\n        code_block: |`;
            response.code_block.split('\n').forEach(line => {
                additionalFields += `\n          ${line}`;
            });
        });
    }
    
    return yaml + additionalFields;
}

if (require.main === module) {
    main();
}