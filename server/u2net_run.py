import sys
sys.path.insert(0, 'U-2-Net')

import io as iom

from skimage import transform, io
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

model_dir = './U-2-Net/saved_models/' + model_name + '/' + model_name + '.pth'

def normPRED(d):
    ma = torch.max(d)
    mi = torch.min(d)

    dn = (d-mi)/(ma-mi)

    return dn

# Open image 
buffer = iom.BytesIO()
buffer.write(sys.stdin.buffer.read())
buffer.seek(0)

orig_image = Image.open(buffer)
image = np.array(orig_image)


# Load u2net
if(model_name=='u2net'):
    net = U2NET(3,1)
elif(model_name=='u2netp'):
    net = U2NETP(3, 1)
if not torch.cuda.is_available():
    net.load_state_dict(torch.load(model_dir, map_location=torch.device('cpu')))
else:
    net.load_state_dict(torch.load(model_dir))
if torch.cuda.is_available():
    net.cuda()
net.eval()

# Preprocess image
label_3 = np.zeros(image.shape)

label = np.zeros(label_3.shape[0:2])
if(3==len(label_3.shape)):
    label = label_3[:,:,0]
elif(2==len(label_3.shape)):
    label = label_3

if(3==len(image.shape) and 2==len(label.shape)):
    label = label[:,:,np.newaxis]
elif(2==len(image.shape) and 2==len(label.shape)):
    image = image[:,:,np.newaxis]
    label = label[:, :, np.newaxis]
    
transform = transforms.Compose([RescaleT(320), ToTensorLab(flag=0)])
sample = transform({'imidx': np.array([0]), 'image': image, 'label': label})

# Process image
image_process = sample['image'].unsqueeze(0)
image_process = image_process.type(torch.FloatTensor)

if torch.cuda.is_available():
    image_process = Variable(image_process.cuda())
else:
    image_process = Variable(image_process)

d1,d2,d3,d4,d5,d6,d7= net(image_process)

# normalization
pred = d1[:,0,:,:]
pred = normPRED(pred)

# save mask
predict = pred
predict = predict.squeeze()
predict_np = predict.cpu().data.numpy()

im = Image.fromarray(predict_np*255).convert('RGB')
image = image
imo = im.resize((image.shape[1], image.shape[0]), resample=Image.BILINEAR)


pb_np = np.array(imo)

imo.save('output.png')



