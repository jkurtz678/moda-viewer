import webview
import sys
import signal

# allows keybaord interrupts to work
signal.signal(signal.SIGINT, signal.SIG_DFL)

# run webview
plaque_url = sys.argv[1]
webview.create_window('MoDA Plaque', plaque_url)
webview.start() 