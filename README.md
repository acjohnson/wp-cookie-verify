wp-cookie-verify
================

Simple Wordpress auth endpoint used with nginx `ngx_http_auth_request_module`

With this endpoint you can set up a `/auth` location in `nginx.conf` for use with `auth_request` in another location directive

*Note*:
Requires valid cookie strings (`wordpress_logged_in_*`) to be added to a redis instance, recommended to use a TTL for each cookie maybe 2 weeks or less depending on your use-case...

As long as the cookie is in redis the `cookie` header will be passed to the `/auth` endpoint and wp-cookie-verify will determine if it is legit.

Example
=======
`nginx.conf`
```ini
        location / {
        ...     # Use case-insensitive regex nested location to determine where authentication is required...
                location ~* .*(private|bill-pay|wp-json|wp-admin).* {
                        error_page 401 = @error401;
                        auth_request /auth;
                        proxy_set_header cookie $http_cookie;
                        proxy_pass_request_headers on;

                        proxy_set_header Host $http_host;
                        proxy_set_header X-Forwarded-Proto $scheme;
                        proxy_set_header X-Original-Forwarded-For $http_x_forwarded_for;
                        proxy_set_header Proxy "";
                        proxy_redirect off;

                        proxy_pass http://wp_upstream;
                }
        }

        # Set up auth endpoint
        location = /auth {
                proxy_pass_request_body off;
                proxy_set_header Content-Length "";

                proxy_set_header X-Original-Host $host;
                proxy_set_header X-Original-URI $request_uri;
                proxy_set_header X-Request-ID $req_id;

                proxy_set_header Host <wp-cookie-verify_fqdn>;
                proxy_redirect off;
                proxy_pass http://<wp-cookie-verify_fqdn>:8081/auth;
        }

        location @error401 {
                # Redirect to Wordpress /logout assuming /logout automatically logs out the user without a prompt...
                return 302 https://$host/logout;
        }
```
