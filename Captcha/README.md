# 智能验证码识别程序

一个基于深度学习的智能验证码识别程序，支持多种类型验证码的生成和识别。

## 功能特性

### 1. 多种验证码类型支持
- **basic**: 基础验证码，简单文字显示
- **noisy**: 带噪点和干扰线的验证码（默认）
- **gradient**: 渐变背景验证码
- **checkerboard**: 棋盘格背景验证码
- **chinese**: 中文验证码
- **math**: 数学运算验证码（如 12+34=?）
- **blur**: 模糊效果验证码
- **colorful**: 彩色字符验证码

### 2. 训练功能
- 基于 PyTorch 的深度学习模型
- **CPU 训练模式**（无需显卡）
- 自动划分训练/验证集
- 保存最佳模型

### 3. 识别功能
- 支持单张图片识别
- 实时预测结果

## 项目结构

```
Captcha/
├── captcha_types.py    # 验证码生成器（支持8种类型）
├── config.py           # 配置文件
├── dataset.py          # 数据集处理
├── generate_captcha.py # 数据集生成
├── model.py            # CNN模型
├── train.py            # 训练脚本（CPU模式）
├── predict.py          # 预测脚本
├── test_captcha_types.py # 测试脚本
├── main.py             # 主入口
└── requirements.txt    # 依赖列表
```

## 安装依赖

```bash
# 创建虚拟环境
python3 -m venv venv
source venv/bin/activate

# 安装依赖
pip install -r requirements.txt
```

## 使用方法

### 1. 生成训练数据集

```bash
# 生成默认类型（noisy）的数据集
python main.py generate --train 10000 --test 2000

# 指定验证码类型
python main.py generate --train 5000 --test 1000 --type gradient

# 支持的类型：basic, noisy, gradient, checkerboard, chinese, math, blur, colorful, mixed
```

### 2. 训练模型

```bash
python main.py train
```

训练将在 CPU 上进行，自动保存最佳模型到 `models/captcha_model.pth`。

### 3. 预测验证码

```bash
python main.py predict path/to/captcha.png
```

### 4. 测试验证码生成

```bash
# 测试所有类型（各生成一张）
python main.py test --mode test

# 生成测试数据集
python main.py test --mode generate --num 50
```

## 模型架构

采用卷积神经网络（CNN）架构：

```
输入层 (1, 40, 120)
    ↓
卷积层1: Conv2d(1, 32, 3) → ReLU → MaxPool2d
    ↓
卷积层2: Conv2d(32, 64, 3) → ReLU → MaxPool2d
    ↓
卷积层3: Conv2d(64, 128, 3) → ReLU → MaxPool2d
    ↓
卷积层4: Conv2d(128, 256, 3) → ReLU → MaxPool2d
    ↓
全连接层1: Linear(256×7×2, 1024) → ReLU → Dropout
    ↓
全连接层2: Linear(1024, 4×36)
    ↓
输出层 (4, 36) → 4位字符 × 36种字符(0-9, A-Z)
```

## 配置参数

在 `config.py` 中可修改以下参数：

| 参数 | 默认值 | 说明 |
|------|--------|------|
| CHAR_SET | '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ' | 字符集 |
| IMAGE_WIDTH | 120 | 图片宽度 |
| IMAGE_HEIGHT | 40 | 图片高度 |
| MAX_CAPTCHA | 4 | 验证码位数 |
| BATCH_SIZE | 64 | 批次大小 |
| EPOCHS | 100 | 训练轮数 |
| LEARNING_RATE | 0.001 | 学习率 |

## 技术栈

- Python 3.8+
- PyTorch 2.5+
- PIL/Pillow
- NumPy
- scikit-learn
- tqdm

## 许可证

MIT License