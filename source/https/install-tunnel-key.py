from pathlib import Path
import shutil

directory = Path.home() / '.ssh'
directory.mkdir(mode=0o700, exist_ok=True)
target = directory / 'authorized_keys'
backup = directory / 'authorized_keys.before-vohive-https-20261002'
public = Path('/tmp/vohive-https-stage/tunnel.openssh.pub').read_text().strip()
existing = target.read_text() if target.exists() else ''
if public.split()[1] in existing:
    raise SystemExit('Tunnel key already present.')
if target.exists() and not backup.exists():
    shutil.copy2(target, backup)
options = ('restrict,port-forwarding,permitopen="127.0.0.1:17000",'
           'permitlisten="127.0.0.1:17575",command="/bin/false"')
target.write_text(existing.rstrip() + '\n' + options + ' ' + public + '\n')
target.chmod(0o600)
print('Restricted tunnel key installed.')
