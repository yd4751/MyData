import os
from captcha_types import CaptchaGenerator
from config import OUTPUT_DIR

def test_all_captcha_types():
    generator = CaptchaGenerator()
    captcha_types = generator.get_available_types()
    
    print("Testing all available captcha types:")
    print("=" * 50)
    
    for captcha_type in captcha_types:
        print(f"\nGenerating {captcha_type} captcha...")
        try:
            image, text = generator.generate(captcha_type)
            
            output_path = os.path.join(OUTPUT_DIR, f'{captcha_type}_{text}.png')
            image.save(output_path)
            
            print(f"  Success! Text: {text}, Saved to: {output_path}")
            
            image.show(title=f'{captcha_type} captcha: {text}')
            
        except Exception as e:
            print(f"  Failed: {e}")
    
    print("\n" + "=" * 50)
    print("Test completed! Check the output directory for generated images.")

def generate_dataset_with_types(output_dir, num_per_type=100):
    generator = CaptchaGenerator()
    captcha_types = generator.get_available_types()
    
    os.makedirs(output_dir, exist_ok=True)
    
    print(f"Generating {num_per_type} samples for each captcha type...")
    
    for captcha_type in captcha_types:
        type_dir = os.path.join(output_dir, captcha_type)
        os.makedirs(type_dir, exist_ok=True)
        
        for i in range(num_per_type):
            try:
                image, text = generator.generate(captcha_type)
                image.save(os.path.join(type_dir, f'{text}_{i}.png'))
                
                if (i + 1) % 50 == 0:
                    print(f"  {captcha_type}: {i + 1}/{num_per_type}")
                    
            except Exception as e:
                print(f"  Failed to generate {captcha_type} captcha: {e}")
    
    print("Dataset generation complete!")

if __name__ == '__main__':
    import argparse
    
    parser = argparse.ArgumentParser(description='Test various captcha types')
    parser.add_argument('--mode', choices=['test', 'generate'], default='test',
                        help='Mode: test (show one of each type) or generate (create dataset)')
    parser.add_argument('--output', default=OUTPUT_DIR, help='Output directory')
    parser.add_argument('--num', type=int, default=100, help='Number per type for generate mode')
    
    args = parser.parse_args()
    
    if args.mode == 'test':
        test_all_captcha_types()
    else:
        generate_dataset_with_types(args.output, args.num)