# Edge VPS TLS Certificates

Upload the **STAR_functionfly_com** wildcard certificate (and chain) to the edge VPS nodes (**217.160.124.206**, **209.46.125.113**) so `edge.functionfly.com` and other subdomains are served over HTTPS.

## What you need

1. **Certificate files** (from your CA — often as a **zip of four certs**):
   - **Leaf**: `STAR_functionfly_com` (the wildcard cert for `*.functionfly.com`)
   - **Intermediates**: e.g. Sectigo Public Server Authentication, SSL2BUY Domain Validation, USERTrust RSA (the chain your CA provided)
2. **Private key**: The key you used to generate the CSR. CAs usually do **not** put this in the zip — you must add it yourself as `privkey.pem` or a `.key` file.  
   **If your keys are already on each VPS** (e.g. generated there or uploaded earlier), you only need to upload the **certificate chain** (`fullchain.pem`); use `-KeysOnServer` (Windows) or `KEYS_ON_SERVER=1` (bash), or with WinSCP upload only `fullchain.pem`.

All of these can be `.cer`, `.crt`, or `.pem`. The scripts convert as needed.

### If you have a zip from the CA (four certs)

1. **Extract the zip** to a folder (e.g. `Downloads\STAR_functionfly_com` or `Downloads\functionfly-certs`).
2. **If keys are not on the VPS:** Add your private key to that same folder as `privkey.pem` (or the `.key` file from your CSR). The zip typically only has the four certificates.  
   **If keys are already on each VPS (or were generated on the VPS):** You don't need the key locally. Put only the four certs in `certs-in/`, run `./prepare-certs.sh`, then `KEYS_ON_SERVER=1 ./upload-certs.sh`.
3. Use that folder as `-CertFolder` (or `-CertZip` with the zip path) when running the Windows script; or copy the extracted folder into `certs-in/` for the bash script.

## From Windows (certificates in Downloads)

If the cert files are on a **Windows** machine (e.g. `Downloads\STAR_functionfly_com`):

1. **Install (if needed):**
   - **Git for Windows** (includes OpenSSL): https://git-scm.com/download/win  
   - **OpenSSH Client** (usually already on Windows 10/11): Settings → Apps → Optional features → “OpenSSH Client”

2. **If keys are not on the VPS:** Put your private key in the same folder as the certs (e.g. `STAR_functionfly_com`) as `privkey.pem` or any `.key` file.  
   **If keys are already on each VPS:** Omit the key and use `-KeysOnServer` so only the chain is uploaded.

3. **Open PowerShell**, go to the edge deploy folder, and run (use your actual path to the **extracted** cert folder or the zip):

   ```powershell
   cd path\to\functionfly\deploy\edge
   # Zip (four certs only) — keys already on each VPS:
   .\upload-certs-windows.ps1 -CertZip "$env:USERPROFILE\Downloads\functionfly-certs.zip" -KeysOnServer
   # Or extracted folder with privkey.pem in it:
   .\upload-certs-windows.ps1 -CertFolder "$env:USERPROFILE\Downloads\STAR_functionfly_com"
   # Zip with key inside or next to zip:
   .\upload-certs-windows.ps1 -CertZip "$env:USERPROFILE\Downloads\functionfly-certs.zip"
   ```

   To use a different SSH user (default is `root`):

   ```powershell
   .\upload-certs-windows.ps1 -CertFolder "C:\Users\YourName\Downloads\STAR_functionfly_com" -SshUser root
   ```

4. The script builds the chain and uploads **fullchain.pem** (and **privkey.pem** unless you used `-KeysOnServer`) to both VPS nodes into `/etc/ssl/functionfly/`. You may be prompted for the SSH password (or use key-based auth).

5. **One-time on each VPS:** install Caddy and use the certs (see “Enable TLS on each VPS” below). After that, when you run the script again for renewals, just restart Caddy on each server.

### Using WinSCP (GUI upload)

If you prefer **WinSCP** instead of the PowerShell script:

1. **Prepare fullchain.pem on Windows** (and privkey.pem only if keys are not already on the VPS):
   - Run the PowerShell script once (e.g. `-CertZip path\to\certs.zip -KeysOnServer`); it creates `deploy\edge\certs-out\fullchain.pem`. Use that.
   - Or with a key: script creates both `fullchain.pem` and `privkey.pem` in `certs-out\`.

2. **Install WinSCP**: https://winscp.net/eng/download.php

3. **Connect to each VPS:**
   - **File protocol:** SFTP  
   - **Host:** `217.160.124.206` (then repeat for `209.46.125.113`)  
   - **Port:** 22  
   - **User:** `root` (or your SSH user)  
   - **Password:** (or use key file in Advanced → SSH → Authentication)

4. **On the server:** In the right (remote) panel, go to `/etc/ssl/`. If `functionfly` doesn't exist, right-click → New → Directory → name it `functionfly`. Open that folder.

5. **Upload:** Drag **fullchain.pem** into `/etc/ssl/functionfly/`. If keys are not on the VPS, also drag **privkey.pem** there.

6. **Permissions (on the server):** Set `fullchain.pem` to `644`. If you uploaded **privkey.pem**, right-click it → Properties → Permissions → `600`.  
   (Or via SSH: `chmod 644 /etc/ssl/functionfly/fullchain.pem` and, if needed, `chmod 600 /etc/ssl/functionfly/privkey.pem`.)

7. **Repeat** steps 3–6 for the second server: **209.46.125.113**.

8. Restart Caddy on each node when done (or run the one-time Caddy setup; see "Enable TLS on each VPS" below).

---

## Steps (Linux / WSL)

### 1. Prepare certs locally

Copy your certificate folder (e.g. from Windows `Downloads\STAR_functionfly_com`) into this directory as `certs-in/`:

```bash
cd deploy/edge
mkdir -p certs-in
# Copy into certs-in:
#   - STAR_functionfly_com.cer (or .crt / .pem)
#   - Sectigo....cer, SSL2BUY....cer, USERTrust....cer (or .pem)
#   - privkey.pem (or your .key file from the CA)
```

If the private key is in a different format (e.g. from the CA portal), ensure it’s PEM. Convert if needed:

```bash
openssl rsa -in yourkey.key -out certs-in/privkey.pem
```

Build the chain and key output:

```bash
./prepare-certs.sh
```

This produces `certs-out/fullchain.pem` and `certs-out/privkey.pem`.

### 2. Upload to VPS nodes

From the same machine (with SSH access to the edge nodes):

```bash
./upload-certs.sh [ssh_user]
```

Default `ssh_user` is `root`. Nodes are defined in the script: **217.160.124.206**, **209.46.125.113**. Certs are placed under `/etc/ssl/functionfly/` on each node.

### 3. Enable TLS on each VPS (one-time)

On each node, install Caddy and point it at the uploaded certs so it terminates TLS and proxies to the FunctionFly Edge proxy on port 8080:

```bash
# From your machine:
scp deploy/edge/setup-edge-tls.sh deploy/edge/Caddyfile.edge root@217.160.124.206:/tmp/
scp deploy/edge/setup-edge-tls.sh deploy/edge/Caddyfile.edge root@209.46.125.113:/tmp/

# On each node:
ssh root@217.160.124.206 'bash /tmp/setup-edge-tls.sh'
ssh root@209.46.125.113 'bash /tmp/setup-edge-tls.sh'
```

Caddy will listen on 80/443 and proxy to `localhost:8080` (the existing `functionfly-edge` service).

### 4. Renewals

When the wildcard cert is renewed, repeat steps 1–2, then on each node:

```bash
sudo systemctl restart caddy
```

## Files

| File | Purpose |
|------|--------|
| `upload-certs-windows.ps1` | **Windows:** one script to build chain + upload to 217.160.124.206 and 209.46.125.113 |
| `prepare-certs.sh` | (Linux/WSL) Build `fullchain.pem` + `privkey.pem` from `certs-in/` |
| `upload-certs.sh` | (Linux/WSL) SCP certs to both edge VPS nodes |
| `Caddyfile.edge` | Caddy config: TLS + reverse proxy to :8080 |
| `setup-edge-tls.sh` | Run on each VPS: install Caddy, deploy Caddyfile |

## Security

- `certs-out/` and `certs-in/` contain secrets; they are gitignored. Do not commit private keys or certs.
- On the VPS, `/etc/ssl/functionfly/privkey.pem` is mode `600`.
