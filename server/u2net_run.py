import sys
sys.path.insert(0, 'U-2-Net')

import io

from skimage import transform
import torch
import torchvision
from torch.autograd import Variable
import torch.nn as nn
import torch.nn.functional as F
from torch.utils.data import Dataset, DataLoader
from torchvision import transforms

from data_loader import RescaleT
from data_loader import ToTensor
from data_loader import ToTensorLab
from data_loader import SalObjDataset

import numpy as np
from PIL import Image

from model import U2NET # full size version 173.6 MB
from model import U2NETP # small version u2net 4.7 MB

model_name='u2net'#u2netp

model_dir = './U-2-Net/saved_models/'+ model_name + '/' + model_name + '.pth'


def normPRED(d):
    ma = torch.max(d)
    mi = torch.min(d)

    dn = (d-mi)/(ma-mi)

    return dn

buffer = io.StringIO.StringIO()
buffer.write(sys.stdin.read())
buffer.seek(0)

img = Image.open(buffer)
print(img)

print("test")




