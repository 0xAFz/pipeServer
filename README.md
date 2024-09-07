# Pipe

Pipe is a Telegram Mini App with E2EE (ECC + AES), Users can send hidden message to each other. 

## Local Development

### Requirements
- [Golang](https://go.dev/doc/install) v1.22.5
- [Redis](https://redis.io/) v7.4
- [Apache Cassandra](https://cassandra.apache.org) v5.0-rc1
- [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/) plugin

### Setting Up Local Development Environment

1. Clone the Pipe git repository:
   ```bash
   git clone https://github.com/0xAFz/pipeServer.git
   cd pipeServer
   ```

2. Start the Cassandra and Redis instance using Docker Compose:
   ```bash
   docker compose -f compose.yml up -d
   ```

3. Set up the environment variables:
   ```bash
   cp .env.example .env
   ```
   Edit the `.env` file and replace the placeholder values with your actual configuration.

4. Run the application:
   ```bash
   make run
   # or
   go run main.go
   ```

## Production Deployment

### Requirements
- [Nginx](https://nginx.org)
- [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/) plugin

### Deploying Pipe Server in Production

1. Clone the Pipe git repository on your production server:
   ```bash
   git clone https://github.com/0xAFz/pipeServer.git
   cd pipeServer
   ```
2. Set up the environment variables:
   ```bash
   cp .env.example .env
   ```
   Edit the `.env` file and replace the placeholder values with your actual configuration.

3. Build and start the production services:
   ```bash
   docker compose -f prod.compose.yml up --build -d
   ```

   At this point, the Pipe server will be running on `http://127.0.0.1:1323`, accessible only from the local network.

4. Set up Nginx as a reverse proxy to route requests to the Pipe service. We'll cover the process of installing Nginx directly on the VM.

### Setting Up HTTPS with Nginx

1. Install Nginx:
   ```bash
   # Debian/Ubuntu
   sudo apt update
   sudo apt install nginx -y

   # CentOS/RHEL
   sudo dnf install nginx -y
   # or
   sudo yum install nginx -y

   # FreeBSD
   doas pkg install nginx
   ```

2. Generate SSL/TLS certificates for your domain (e.g., api.domain.tld):
   ```bash
   # Using Let's Encrypt (certbot)
   sudo certbot certonly --nginx -d api.domain.tld
   ```

3. Disable the default Nginx configuration:
   ```bash
   sudo unlink /etc/nginx/sites-enabled/default
   ```

4. Create a new Nginx configuration for Pipe:
   ```bash
   sudo cp /path/to/pipeServer/nginx.conf.d/nginx.conf /etc/nginx/sites-available/pipe
   ```

5. Edit the Nginx configuration file:
   ```bash
   sudo nano /etc/nginx/sites-available/pipe
   ```
   Replace `api.domain.tld` with your actual domain and update the SSL certificate paths.

6. Enable the new configuration:
   ```bash
   sudo ln -s /etc/nginx/sites-available/pipe /etc/nginx/sites-enabled
   ```

7. Test the Nginx configuration:
   ```bash
   sudo nginx -t
   ```

8. If the test is successful, reload Nginx:
   ```bash
   sudo systemctl reload nginx
   ```

### Verifying the Deployment

Pipe server should now be accessible at `https://api.domain.tld`. To verify:

```bash
curl -s https://api.domain.tld
```

You should receive a response like this:
```json
{
  "status": "ok"
}
```

## Maintenance and Monitoring

1. Regularly update your server and Docker images:
   ```bash
   sudo apt update && sudo apt upgrade -y
   docker compose -f prod.compose.yml pull
   docker compose -f prod.compose.yml up -d
   ```

2. Monitor logs:
   ```bash
   docker compose -f prod.compose.yml logs -f
   ```

3. Set up a monitoring solution like Prometheus and Grafana for better observability.

4. Implement regular backups of your Cassandra data.

## Troubleshooting

- If you encounter issues, check the Docker logs:
  ```bash
  docker compose -f prod.compose.yml logs
  ```

- Verify Nginx logs:
  ```bash
  sudo tail -f /var/log/nginx/access.log
  sudo tail -f /var/log/nginx/error.log
  ```

- Ensure all required ports are open in your firewall settings.

Remember to keep your production environment secure by following best practices for server hardening and regularly updating all components.