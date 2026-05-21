import torch
import numpy as np
from PIL import Image
import os

from model import load_model
from dataset import decode_label
from config import MODEL_DIR, IMAGE_WIDTH, IMAGE_HEIGHT

def predict(image_path, model_path=None):
    if model_path is None:
        model_path = os.path.join(MODEL_DIR, 'captcha_model.pth')
    
    model = load_model(model_path)
    device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')
    model.to(device)
    model.eval()
    
    image = Image.open(image_path).convert('L')
    image = image.resize((IMAGE_WIDTH, IMAGE_HEIGHT))
    image = np.array(image) / 255.0
    image = image[np.newaxis, np.newaxis, :, :]
    
    with torch.no_grad():
        image_tensor = torch.from_numpy(image).to(device).float()
        output = model(image_tensor)
        pred = decode_label(output[0].cpu().detach().numpy())
    
    return pred

if __name__ == '__main__':
    import sys
    if len(sys.argv) != 2:
        print('Usage: python predict.py <image_path>')
        sys.exit(1)
    
    image_path = sys.argv[1]
    result = predict(image_path)
    print(f'Predicted captcha: {result}')