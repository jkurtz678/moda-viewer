#
#   Hello World client in Python
#   Connects REQ socket to tcp://localhost:5555
#   Sends "Hello" to server, expects "World" back
#

import zmq

context = zmq.Context()

#  Socket to talk to server
print("Connecting to hello world server…")
socket = context.socket(zmq.REQ)
socket.connect("tcp://localhost:5555")
socket.send(b"https://www.google.com/maps/place/Sichuan,+China/@29.5851258,100.2210999,7z/data=!4m5!3m4!1s0x36e4e73368bdcdb3:0xde8f7ccf8f99feb9!8m2!3d30.6508899!4d104.07572")

#  Get the reply.
message = socket.recv()
print("Received reply %s" % message)