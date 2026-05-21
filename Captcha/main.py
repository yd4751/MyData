import argparse
import os

def main():
    parser = argparse.ArgumentParser(description='智能验证码识别程序')
    subparsers = parser.add_subparsers(dest='command', help='可用命令')
    
    gen_parser = subparsers.add_parser('generate', help='生成验证码数据集')
    gen_parser.add_argument('--train', type=int, default=10000, help='训练集数量')
    gen_parser.add_argument('--test', type=int, default=2000, help='测试集数量')
    gen_parser.add_argument('--type', choices=['basic', 'noisy', 'gradient', 'checkerboard', 'chinese', 'math', 'blur', 'colorful', 'mixed'], 
                           default='noisy', help='验证码类型')
    
    test_parser = subparsers.add_parser('test', help='测试各种验证码类型')
    test_parser.add_argument('--mode', choices=['test', 'generate'], default='test',
                            help='test:展示每种类型一个, generate:生成数据集')
    test_parser.add_argument('--num', type=int, default=100, help='每种类型生成数量')
    
    train_parser = subparsers.add_parser('train', help='训练模型')
    
    predict_parser = subparsers.add_parser('predict', help='预测验证码')
    predict_parser.add_argument('image_path', help='验证码图片路径')
    
    args = parser.parse_args()
    
    if args.command == 'generate':
        from generate_captcha import generate_dataset
        generate_dataset(args.train, args.test, args.type)
    elif args.command == 'test':
        from test_captcha_types import test_all_captcha_types, generate_dataset_with_types
        if args.mode == 'test':
            test_all_captcha_types()
        else:
            from config import OUTPUT_DIR
            generate_dataset_with_types(OUTPUT_DIR, args.num)
    elif args.command == 'train':
        from train import train
        train()
    elif args.command == 'predict':
        from predict import predict
        result = predict(args.image_path)
        print(f'识别结果: {result}')
    else:
        parser.print_help()

if __name__ == '__main__':
    main()