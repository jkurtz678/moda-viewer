import webview
import sys
import zmq
import threading

def zmq_listener(window): 
    context = zmq.Context()
    socket = context.socket(zmq.REP)
    socket.bind("tcp://*:5555")
    try: 
        while True: 
            message = socket.recv()
            new_url = message.decode("utf-8")
            print("Received request: %s" % new_url, flush=True)
            try: 
                window.load_url(new_url)
                socket.send(b"done")
                print("successfully loaded url", flush=True)
            except Exception as e:
                print("Error loading url %s" % e, flush=True)
                socket.send(b"fail")
    except KeyboardInterrupt:
        print('zeromq interrupted!')

# run webview

try: 
    plaque_url = sys.argv[1]
    window = webview.create_window('MoDA Plaque', plaque_url)
    webview.start(zmq_listener, window)
except KeyboardInterrupt:
    print('webview interrupted!')