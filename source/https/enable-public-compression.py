from pathlib import Path
import shutil
import subprocess

site = Path('/etc/nginx/sites-available/tgapi')
backup = site.with_name('tgapi.before-smart-https-20261002')
text = site.read_text()
if backup.exists():
    raise SystemExit('Performance backup already exists; inspect before applying again.')
needle = '        proxy_pass http://127.0.0.1:17575;'
if text.count(needle) != 1:
    raise SystemExit('Unexpected VoHive proxy layout.')
shutil.copy2(site, backup)
text = text.replace(needle, '''        gzip on;
        gzip_comp_level 3;
        gzip_min_length 1024;
        gzip_vary on;
        gzip_proxied any;
        gzip_types text/css application/javascript text/javascript application/json image/svg+xml;
        add_header X-VoHive-Route public always;
''' + needle, 1)
if '--with-http_v2_module' in subprocess.run(['nginx', '-V'], capture_output=True, text=True).stderr:
    text = text.replace('listen 443 ssl;', 'listen 443 ssl http2;', 1)
site.write_text(text)
try:
    subprocess.run(['nginx', '-t'], check=True)
except Exception:
    shutil.copy2(backup, site)
    raise
subprocess.run(['systemctl', 'reload', 'nginx'], check=True)
