#!/bin/bash

# Script to generate Postman collection from Swagger/OpenAPI specification

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SWAGGER_FILE="$PROJECT_ROOT/src/docs/swagger.json"
OUTPUT_FILE="$PROJECT_ROOT/docs/Parallels_Desktop_API.postman_collection.json"

# Check if swagger file exists
if [ ! -f "$SWAGGER_FILE" ]; then
    echo "Error: Swagger file not found at $SWAGGER_FILE"
    echo "Please run 'make generate-swagger' first"
    exit 1
fi

# Check if npx is available
if ! command -v npx &> /dev/null; then
    echo "Error: npx is not installed. Please install Node.js and npm"
    exit 1
fi

echo "Generating Postman collection from Swagger/OpenAPI specification..."

# Use npx to run openapi-to-postmanv2 without requiring global installation
npx --yes openapi-to-postmanv2 \
    -s "$SWAGGER_FILE" \
    -o "$OUTPUT_FILE" \
    -p \
    -O folderStrategy=Tags,requestParametersResolution=Example,exampleParametersResolution=Example,indentCharacter=Tab

if [ $? -eq 0 ]; then
    echo "✓ Postman collection generated successfully at: $OUTPUT_FILE"
    
    # Post-process to replace <string> placeholders with Postman variables
    echo "Post-processing collection to add Postman variables..."
    
    if command -v node &> /dev/null; then
        # Create temporary post-processing script
        TEMP_SCRIPT=$(mktemp /tmp/postman-process.XXXXXX.js)
        cat > "$TEMP_SCRIPT" << 'EOF'
const fs = require('fs');
const inputFile = process.argv[2];
const collection = JSON.parse(fs.readFileSync(inputFile, 'utf8'));

// Replace baseUrl with protocol and host, and add /api to path
const traverse = (obj) => {
    if (typeof obj !== 'object' || obj === null) return;
    if (obj.host && Array.isArray(obj.host) && obj.host[0] === '{{baseUrl}}') {
        obj.host = ['{{host}}'];
        obj.protocol = '{{protocol}}';
        if (obj.path && Array.isArray(obj.path) && obj.path[0] !== 'api') {
            obj.path.unshift('api');
        }
    }
    for (let k in obj) {
        traverse(obj[k]);
    }
};
traverse(collection);

// Replace API Key auth with Bearer token auth
const replaceAuth = (obj) => {
    if (typeof obj !== 'object' || obj === null) return;
    if (obj.auth && obj.auth.type === 'apikey') {
        obj.auth = {
            type: 'bearer',
            bearer: [{ key: 'token', value: '{{token}}', type: 'string' }]
        };
    }
    if (obj.header && Array.isArray(obj.header)) {
        obj.header = obj.header.filter(h => h.key !== 'X-Api-Key');
    }
    for (let k in obj) {
        if (typeof obj[k] === 'object') {
            replaceAuth(obj[k]);
        }
    }
};
replaceAuth(collection);

// Add test script to auth endpoint
const addAuthTest = (items) => {
    if (!items) return;
    for (let item of items) {
        if (item.item) {
            addAuthTest(item.item);
        } else if (item.request && item.request.url) {
            const path = item.request.url.path || [];
            if (path.includes('auth') && path.includes('token') && item.request.method === 'POST') {
                if (item.name && item.name.toLowerCase().includes('generate')) {
                    item.event = [{
                        listen: 'test',
                        script: {
                            type: 'text/javascript',
                            exec: [
                                'var jsonData = pm.response.json();',
                                'if (jsonData.token) {',
                                '    pm.environment.set("token", jsonData.token);',
                                '    console.log("Token saved to environment:", jsonData.token);',
                                '}'
                            ]
                        }
                    }];
                }
            }
        }
    }
};
addAuthTest(collection.item);

// Add variables
collection.variable = [
    { key: 'baseUrl', value: '{{protocol}}://{{host}}/api' },
    { key: 'protocol', value: 'http' },
    { key: 'host', value: 'localhost' },
    { key: 'USER-EMAIL', value: 'root@localhost' },
    { key: 'USER-PASSWORD', value: '' },
    { key: 'token', value: '' }
];

// String replacements
let jsonStr = JSON.stringify(collection, null, 4);
jsonStr = jsonStr
    .replace(/\\n\\t\\"email\\":\s*\\"<string>\\"/g, '\\n\\t\\"email\\": \\"{{USER-EMAIL}}\\"')
    .replace(/\\n\\t\\"password\\":\s*\\"<string>\\"/g, '\\n\\t\\"password\\": \\"{{USER-PASSWORD}}\\"');

fs.writeFileSync(inputFile, jsonStr);
EOF
        
        node "$TEMP_SCRIPT" "$OUTPUT_FILE"
        rm "$TEMP_SCRIPT"
        echo "✓ Post-processing complete"
    fi
else
    echo "✗ Failed to generate Postman collection"
    exit 1
fi
