import os

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
DATA_DIR = os.path.join(BASE_DIR, 'data')
TRAIN_DIR = os.path.join(DATA_DIR, 'train')
TEST_DIR = os.path.join(DATA_DIR, 'test')
MODEL_DIR = os.path.join(BASE_DIR, 'models')
OUTPUT_DIR = os.path.join(BASE_DIR, 'output')

os.makedirs(DATA_DIR, exist_ok=True)
os.makedirs(TRAIN_DIR, exist_ok=True)
os.makedirs(TEST_DIR, exist_ok=True)
os.makedirs(MODEL_DIR, exist_ok=True)
os.makedirs(OUTPUT_DIR, exist_ok=True)

CHAR_SET = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ'
CHAR_COUNT = len(CHAR_SET)
IMAGE_WIDTH = 120
IMAGE_HEIGHT = 40
MAX_CAPTCHA = 4

BATCH_SIZE = 64
EPOCHS = 100
LEARNING_RATE = 0.001