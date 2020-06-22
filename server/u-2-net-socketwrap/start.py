from socket import *
from PIL import Image
import fileinput
import io
import struct

import u2net

host = 'localhost'
port = 8800
addr = (host,port)

tcp_socket = socket(AF_INET, SOCK_STREAM)
tcp_socket.bind(addr)
tcp_socket.listen(1)

while True:
    conn, addr = tcp_socket.accept()

    #   read file size
    data = conn.recv(4)
    packet_length = int.from_bytes(data, 'big')

    #   read file
    data = conn.recv(packet_length)
    image = Image.open(io.BytesIO(bytearray(data)))

    res = u2net.run(image)

    #   send res back
    buff = io.BytesIO()
    res.save(buff, 'PNG')

    conn.send(struct.pack('>I', buff.getbuffer().nbytes) + buff.getbuffer())
    conn.close()

tcp_socket.close()