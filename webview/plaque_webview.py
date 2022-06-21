import webview
import sys
import signal

class Api():
    def toggleFullscreen(self):
        webview.windows[0].toggle_fullscreen()
    def setTitle(self, title):
        webview.windows[0].set_title(title)

# allows keybaord interrupts to work
signal.signal(signal.SIGINT, signal.SIG_DFL)

# run webview
plaque_url = sys.argv[1]
api = Api()
webview.create_window('MoDA Plaque', plaque_url, js_api=api)
webview.start(debug=True) 