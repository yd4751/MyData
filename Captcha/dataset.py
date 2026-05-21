import os
import numpy as np
from PIL import Image
from torch.utils.data import Dataset
from config import CHAR_SET, MAX_CAPTCHA, IMAGE_WIDTH, IMAGE_HEIGHT, TRAIN_DIR, TEST_DIR

class CaptchaDataset(Dataset):
    def __init__(self, data_dir, transform=None):
        self.data_dir = data_dir
        self.transform = transform
        self.image_files = [f for f in os.listdir(data_dir) if f.endswith('.png') or f.endswith('.jpg')]
    
    def __len__(self):
        return len(self.image_files)
    
    def __getitem__(self, idx):
        img_name = self.image_files[idx]
        img_path = os.path.join(self.data_dir, img_name)
        label = img_name.split('.')[0]
        
        image = Image.open(img_path).convert('L')
        image = image.resize((IMAGE_WIDTH, IMAGE_HEIGHT))
        image = np.array(image) / 255.0
        image = image[np.newaxis, :, :]
        
        if self.transform:
            image = self.transform(image)
        
        label_tensor = self.encode_label(label)
        return image, label_tensor
    
    def encode_label(self, label):
        one_hot = np.zeros((MAX_CAPTCHA, len(CHAR_SET)), dtype=np.float32)
        for i, char in enumerate(label[:MAX_CAPTCHA]):
            if char in CHAR_SET:
                one_hot[i, CHAR_SET.index(char)] = 1.0
        return one_hot

def load_train_dataset():
    return CaptchaDataset(TRAIN_DIR)

def load_test_dataset():
    return CaptchaDataset(TEST_DIR)

def decode_label(one_hot):
    label = ''
    for i in range(MAX_CAPTCHA):
        idx = np.argmax(one_hot[i])
        label += CHAR_SET[idx]
    return label