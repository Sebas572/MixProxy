<div style="display: flex; justify-content: space-between; align-items: flex-end;">

<h1 style="margin-bottom: .2rem;">MixProxy</h1>

[![Go Version](https://img.shields.io/badge/Go-1.24.3-blue.svg)](https://golang.org/)

</div>

MixProxy is a high-performance reverse proxy server written in Go, designed to efficiently route traffic to services with load balancing capabilities. It includes an administration panel for easy configuration and monitoring.

## Features

- **Load balancing**: intelligent distribution of traffic across multiple servers based on capacity weights
- **SSL/TLS support**: automatic HTTPS redirection and SSL certificate management
- **Subdomain routing**: subdomain-based traffic routing to different backend services.
- **Administration panel**: easy-to-use web interface for configuration and monitoring.
- **Docker support**: containerized deployment with Docker and Docker Compose.

## Architecture

MixProxy consists of three main components:

1. **Proxy core**: the main reverse proxy engine that handles HTTP/HTTPS traffic, load balancing, and routing.
2. **Management API**: RESTful API for managing proxy configuration and retrieving statistics.
3. **Administration panel**: React-based web interface for administrators to configure and monitor the proxy.

The proxy supports both root domain and subdomain-based routing, with configurable load balancers for each route. It automatically handles SSL termination and can generate self-signed certificates for development.

## Technologies Used

### BackEnd
- **Go**: Core proxy logic and server implementation
- **HTTP/2**: Modern HTTP protocol support

### FrontEnd (Admin Panel)
- **React**: User interface framework
- **TypeScript**: Type-safe JavaScript
- **Vite**: Fast build tool and development server
- **Tailwind CSS**: Utility-first CSS framework
- **shadcn/ui**: Modern UI components
