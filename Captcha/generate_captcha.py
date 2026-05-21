import os
from captcha_types import CaptchaGenerator
from config import TRAIN_DIR, TEST_DIR

def generate_dataset(num_train=10000, num_test=2000, captcha_type='noisy'):
    generator = CaptchaGenerator()
    
    if captcha_type == 'mixed':
        types = generator.get_available_types()
        types.remove('chinese')
        types.remove('math')
    else:
        types = [captcha_type]
    
    print(f"Generating {num_train} training images with type: {captcha_type}")
    for i in range(num_train):
        if captcha_type == 'mixed':
            selected_type = types[i % len(types)]
        else:
            selected_type = captcha_type
        
        try:
            image, text = generator.generate(selected_type)
            image = image.convert('L')
            image.save(os.path.join(TRAIN_DIR, f'{text}_{i}.png'))
            
            if (i + 1) % 1000 == 0:
                print(f'Generated {i + 1}/{num_train} training images')
                
        except Exception as e:
            print(f"Failed to generate training image {i}: {e}")
    
    print(f"\nGenerating {num_test} test images with type: {captcha_type}")
    for i in range(num_test):
        if captcha_type == 'mixed':
            selected_type = types[i % len(types)]
        else:
            selected_type = captcha_type
        
        try:
            image, text = generator.generate(selected_type)
            image = image.convert('L')
            image.save(os.path.join(TEST_DIR, f'{text}_{i}.png'))
            
            if (i + 1) % 500 == 0:
                print(f'Generated {i + 1}/{num_test} test images')
                
        except Exception as e:
            print(f"Failed to generate test image {i}: {e}")
    
    print('\nDataset generation complete!')

if __name__ == '__main__':
    generate_dataset()