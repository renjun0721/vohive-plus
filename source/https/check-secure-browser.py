import json
import subprocess
import time
import urllib.request

import websocket

browser = subprocess.Popen([
    'chromium', '--headless', '--no-sandbox', '--disable-dev-shm-usage',
    '--remote-debugging-address=127.0.0.1', '--remote-debugging-port=9223',
    '--remote-allow-origins=http://127.0.0.1:9223',
    '--user-data-dir=/tmp/vohive-https-browser',
    '--use-fake-ui-for-media-stream', '--use-fake-device-for-media-stream',
    'https://xjp.721609.xyz/',
], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
connection = None
sequence = 0

def request(method, params=None):
    global sequence
    sequence += 1
    connection.send(json.dumps({'id': sequence, 'method': method, 'params': params or {}}))
    while True:
        result = json.loads(connection.recv())
        if result.get('id') == sequence:
            if 'error' in result:
                raise RuntimeError(result['error'])
            return result['result']

def evaluate(expression):
    result = request('Runtime.evaluate', {
        'expression': expression, 'awaitPromise': True, 'returnByValue': True,
    })
    if result.get('exceptionDetails'):
        raise RuntimeError('Browser evaluation failed.')
    return result.get('result', {}).get('value')

try:
    for _ in range(100):
        try:
            with urllib.request.urlopen('http://127.0.0.1:9223/json') as response:
                pages = json.load(response)
            page = next(p for p in pages if p['type'] == 'page')
            connection = websocket.create_connection(
                page['webSocketDebuggerUrl'], timeout=30, origin='http://127.0.0.1:9223',
            )
            break
        except Exception:
            time.sleep(0.2)
    if connection is None:
        raise RuntimeError('Browser did not start.')
    for _ in range(100):
        state = evaluate('({title:document.title, secure:window.isSecureContext, ready:document.readyState})')
        if state.get('title') == 'VoHive Plus' and state.get('ready') == 'complete':
            break
        time.sleep(0.2)
    result = evaluate('''(async () => {
        const stream = await navigator.mediaDevices.getUserMedia({audio:true});
        const result = {title:document.title, secureContext:window.isSecureContext,
                        microphoneAvailable:stream.getAudioTracks().length > 0};
        stream.getTracks().forEach(track => track.stop());
        return result;
    })()''')
    print(json.dumps(result))
    assert result['title'] == 'VoHive Plus'
    assert result['secureContext'] and result['microphoneAvailable']
finally:
    if connection is not None:
        connection.close()
    browser.terminate()
    try:
        browser.wait(timeout=5)
    except subprocess.TimeoutExpired:
        browser.kill()
