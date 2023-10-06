## Simplest Possible Authentication Server for Me
An authentication server compatible with [Caddy](https://caddyserver.com/)'s [`forward_auth`](https://caddyserver.com/docs/caddyfile/directives/forward_auth).

The user is asked for a password and a 6-digit TOTP code (two-factor authentication).

By design, only a single user is supported, and that user can only have a single session at a time.

> [!WARNING]  
> While this server has tests, correctness cannot be proved. Use at your own risk.

### Configuration
Use the following environment variables to configure `spasm`:
- `SPASM_PASS_HASH`: The [bcrypt](https://en.wikipedia.org/wiki/Bcrypt) hash of the password. Can be generated with `mkpasswd -m bcrypt`
- `SPASM_TOTP_KEY`: A secret key in base32. Can be generated with `openssl rand 64 | base32 -w0`
  - This is the key that you can set in your 2FA mobile app, such as [Aegis](https://getaegis.app/).
- `SPASM_ADDRESS`: The address to listen on. Defaults to `localhost:5000`
- `SPASM_COOKIE_NAME`: The name of the session cookie. Recommended to leave empty to use the default (`id`)

Example Caddy configuration:
```
example.com {
  route {
    reverse_proxy /login localhost:5000
    forward_auth localhost:5000 {
      uri /check
    }
    reverse_proxy my-backend.local
  }
}
```

### 2FA
To configure your 2FA mobile app, you need to set the following settings:
- Type/Algorithm: TOTP
- Hash: SHA512
- Period: 10 seconds
- Digits: 6

### How it works
There are two endpoints:
- `login`:
  - GET: serves the login page
  - POST: checks credentials and issues a session cookie if correct, then redirects to the previous page
- `check`:
  - if the credentials are correct, responds with 200 OK
  - else, redirects to the login page, remembering the value of the `X-Forwarded-Uri` header to redirect back to

    > [!NOTE]
    > Caddy sets this header automatically when using `forward_auth`
