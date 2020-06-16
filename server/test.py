import sys
import io
from PIL import Image
import fileinput

buffer = io.BytesIO()
buffer.write(sys.stdin.buffer.read())
buffer.seek(0)

#print(buffer.read())

image = Image.open(buffer)
image.show()