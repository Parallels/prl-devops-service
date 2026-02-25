export interface ReverseProxyConfig {
    enabled: boolean;
    host?: string;
    port?: string;
}

export interface ReverseProxyHostTls {
    enabled?: boolean;
    cert?: string;
    key?: string;
}

export interface ReverseProxyHostCors {
    enabled?: boolean;
    allowed_origins?: string[];
    allowed_methods?: string[];
    allowed_headers?: string[];
}

export interface ReverseProxyHostHttpRoute {
    id?: string;
    path?: string;
    target_vm_id?: string;
    target_host?: string;
    target_port?: string;
    schema?: string;
    pattern?: string;
    request_headers?: Record<string, string>;
    response_headers?: Record<string, string>;
}

export interface ReverseProxyHostHttpRouteCreateRequest {
    path?: string;
    target_vm_id?: string;
    target_host?: string;
    target_port?: string;
    schema?: string;
    pattern?: string;
    request_headers?: Record<string, string>;
    response_headers?: Record<string, string>;
}

export interface ReverseProxyHostTcpRoute {
    id?: string;
    target_port?: string;
    target_host?: string;
    target_vm_id?: string;
}

export interface ReverseProxyHostTcpRouteCreateRequest {
    target_port?: string;
    target_host?: string;
    target_vm_id?: string;
}

export interface ReverseProxyHost {
    id: string;
    host: string;
    port: string;
    tls?: ReverseProxyHostTls;
    cors?: ReverseProxyHostCors;
    http_routes?: ReverseProxyHostHttpRoute[];
    tcp_route?: ReverseProxyHostTcpRoute;
}

export interface ReverseProxyHostCreateRequest {
    host: string;
    port: string;
    tls?: ReverseProxyHostTls;
    cors?: ReverseProxyHostCors;
    http_routes?: ReverseProxyHostHttpRoute[];
    tcp_route?: ReverseProxyHostTcpRoute;
}

export interface ReverseProxyHostUpdateRequest {
    host: string;
    port: string;
    tls?: ReverseProxyHostTls;
    cors?: ReverseProxyHostCors;
}
