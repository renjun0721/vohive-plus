from pathlib import Path
import shutil
import subprocess

site = Path('/etc/nginx/sites-available/tgapi')
backup = site.with_name('tgapi.before-vohive-https-20261002')
extra = Path('/etc/nginx/conf.d/vohive-upgrade.conf')
if backup.exists() or extra.exists():
    raise SystemExit('Existing deployment files found; inspect before applying again.')
text = site.read_text()
if text.count('location / {') != 1 or text.count('    listen 443 ssl;') != 1:
    raise SystemExit('Unexpected Nginx site layout.')
shutil.copy2(site, backup)
text = text.replace('location / {', 'location ~ ^/(bot[0-9]+:|file/bot[0-9]+:) {', 1)
location = '''    location / {
        proxy_pass http://127.0.0.1:17575;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto https;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $vohive_connection_upgrade;
        proxy_buffering off;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }

'''
text = text.replace('    listen 443 ssl;', location + '    listen 443 ssl;', 1)
extra.write_text('map $http_upgrade $vohive_connection_upgrade {\n    default upgrade;\n    "" close;\n}\n')
site.write_text(text)
try:
    subprocess.run(['nginx', '-t'], check=True)
except Exception:
    shutil.copy2(backup, site)
    extra.unlink()
    raise
subprocess.run(['systemctl', 'reload', 'nginx'], check=True)
