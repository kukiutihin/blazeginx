# Gateway

## Pipeline

1. Receive request

2. Route match
   | /service/* -> proxy handler
   | /healthcheck/ -> health handler
   | /metrics/ -> metrics handler, localhost only
   | other  -> static handler if enabled
   | other  -> 404 if static disabled

3. Middleware
   | rate limit
   | timeout
   | log

4. Handler
   | proxy
       | find service
       | send request upstream
       | receive response
       | send response to user
   | static
       | path exists -> return file
       | not exists -> index.html or 404

## Config

- Env (local, dev, prod)
- Port 
- Services:
    - name 
    - url
- Routes:
    - path
    - service 
- rate limit:
    - enable/disable
    - rps
- timeout (seconds):
    - upstream 
    - server 
    - iddle
- static
    - enable/disable
    - root

Use: cleanenv

## Logger

Use: slog
