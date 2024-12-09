---
layout: api
title: General
default_host: http://localhost
api_prefix: /api
is_category_document: true
categories:
    - name: General
endpoints:
    - requires_authorization: true
      category: General
      example_blocks:
        - code_block: "curl --location 'http://localhost/api' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod., "http://localhost/api");
            request.Headers.Add("Authorization", "••••••");
            var response = await client.SendAsync(request);
            response.EnsureSuccessStatusCode();
            var responseString = await response.Content.ReadAsStringAsync();
          title: C#
          language: csharp
        - code_block: |
            package main

            import (
              "fmt"
              "net/http"
              "strings"
              "io"
            )

            func main() {
              url := "http://localhost/api"
              method := ""
              client := &http.Client{}
              req, err := http.NewRequest(method, url, payload)
              if err != nil {
                fmt.Println(err)
                return
              }
              req.Header.Add("Content-Type", "application/json")

              req.Header.Add("Authorization", "••••••")
              res, err := client.Do(req)
              if err != nil {
                fmt.Println(err)
                return
              }
              defer res.Body.Close()
              body, err := io.ReadAll(res.Body)
              if err != nil {
                fmt.Println(err)
                return
              }
              fmt.Println(string(body))
            }
          title: Go
          language: go
    - requires_authorization: true
      category: General
      example_blocks:
        - code_block: "curl --location 'http://localhost/api' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod., "http://localhost/api");
            request.Headers.Add("Authorization", "••••••");
            var response = await client.SendAsync(request);
            response.EnsureSuccessStatusCode();
            var responseString = await response.Content.ReadAsStringAsync();
          title: C#
          language: csharp
        - code_block: |
            package main

            import (
              "fmt"
              "net/http"
              "strings"
              "io"
            )

            func main() {
              url := "http://localhost/api"
              method := ""
              client := &http.Client{}
              req, err := http.NewRequest(method, url, payload)
              if err != nil {
                fmt.Println(err)
                return
              }
              req.Header.Add("Content-Type", "application/json")

              req.Header.Add("Authorization", "••••••")
              res, err := client.Do(req)
              if err != nil {
                fmt.Println(err)
                return
              }
              defer res.Body.Close()
              body, err := io.ReadAll(res.Body)
              if err != nil {
                fmt.Println(err)
                return
              }
              fmt.Println(string(body))
            }
          title: Go
          language: go
    - requires_authorization: true
      category: General
      example_blocks:
        - code_block: "curl --location 'http://localhost/api' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod., "http://localhost/api");
            request.Headers.Add("Authorization", "••••••");
            var response = await client.SendAsync(request);
            response.EnsureSuccessStatusCode();
            var responseString = await response.Content.ReadAsStringAsync();
          title: C#
          language: csharp
        - code_block: |
            package main

            import (
              "fmt"
              "net/http"
              "strings"
              "io"
            )

            func main() {
              url := "http://localhost/api"
              method := ""
              client := &http.Client{}
              req, err := http.NewRequest(method, url, payload)
              if err != nil {
                fmt.Println(err)
                return
              }
              req.Header.Add("Content-Type", "application/json")

              req.Header.Add("Authorization", "••••••")
              res, err := client.Do(req)
              if err != nil {
                fmt.Println(err)
                return
              }
              defer res.Body.Close()
              body, err := io.ReadAll(res.Body)
              if err != nil {
                fmt.Println(err)
                return
              }
              fmt.Println(string(body))
            }
          title: Go
          language: go
    - requires_authorization: true
      category: General
      example_blocks:
        - code_block: "curl --location 'http://localhost/api' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod., "http://localhost/api");
            request.Headers.Add("Authorization", "••••••");
            var response = await client.SendAsync(request);
            response.EnsureSuccessStatusCode();
            var responseString = await response.Content.ReadAsStringAsync();
          title: C#
          language: csharp
        - code_block: |
            package main

            import (
              "fmt"
              "net/http"
              "strings"
              "io"
            )

            func main() {
              url := "http://localhost/api"
              method := ""
              client := &http.Client{}
              req, err := http.NewRequest(method, url, payload)
              if err != nil {
                fmt.Println(err)
                return
              }
              req.Header.Add("Content-Type", "application/json")

              req.Header.Add("Authorization", "••••••")
              res, err := client.Do(req)
              if err != nil {
                fmt.Println(err)
                return
              }
              defer res.Body.Close()
              body, err := io.ReadAll(res.Body)
              if err != nil {
                fmt.Println(err)
                return
              }
              fmt.Println(string(body))
            }
          title: Go
          language: go

---
# Generalendpoints 


