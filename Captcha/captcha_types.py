import os
import random
import string
from PIL import Image, ImageDraw, ImageFont, ImageFilter

CHARS_DIGITS = '0123456789'
CHARS_LETTERS = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz'
CHARS_MIXED = CHARS_DIGITS + CHARS_LETTERS
CHARS_CHINESE = '你我他她它这那上下左右前后大小多少有无来去好坏长短远近高低轻重快慢冷热新旧真假对错是非成败得失'

class CaptchaGenerator:
    def __init__(self, width=120, height=40):
        self.width = width
        self.height = height
        try:
            self.font = ImageFont.truetype('arial.ttf', 28)
        except:
            self.font = ImageFont.load_default()
    
    def get_font_size(self, text):
        bbox = self.font.getbbox(text)
        return (bbox[2] - bbox[0], bbox[3] - bbox[1])
    
    def generate_random_text(self, length=4, char_set=CHARS_MIXED):
        return ''.join(random.choice(char_set) for _ in range(length))
    
    def create_noisy_background(self):
        image = Image.new('RGB', (self.width, self.height), (255, 255, 255))
        draw = ImageDraw.Draw(image)
        for _ in range(random.randint(50, 200)):
            x = random.randint(0, self.width - 1)
            y = random.randint(0, self.height - 1)
            r = random.randint(0, 255)
            g = random.randint(0, 255)
            b = random.randint(0, 255)
            draw.point((x, y), (r, g, b))
        return image
    
    def create_gradient_background(self):
        image = Image.new('RGB', (self.width, self.height))
        for y in range(self.height):
            r = 200 + int(55 * (y / self.height))
            g = 210 + int(45 * (y / self.height))
            b = 220 + int(35 * (y / self.height))
            for x in range(self.width):
                image.putpixel((x, y), (r, g, b))
        return image
    
    def create_checkerboard_background(self):
        image = Image.new('RGB', (self.width, self.height), (255, 255, 255))
        draw = ImageDraw.Draw(image)
        tile_size = 10
        for y in range(0, self.height, tile_size):
            for x in range(0, self.width, tile_size):
                if (x // tile_size + y // tile_size) % 2 == 0:
                    draw.rectangle([x, y, x + tile_size, y + tile_size], fill=(240, 240, 240))
        return image
    
    def add_noise_dots(self, image, count=100):
        draw = ImageDraw.Draw(image)
        for _ in range(count):
            x = random.randint(0, self.width - 1)
            y = random.randint(0, self.height - 1)
            size = random.randint(1, 2)
            draw.ellipse([x, y, x + size, y + size], fill=(random.randint(0, 150), random.randint(0, 150), random.randint(0, 150)))
        return image
    
    def add_noise_lines(self, image, count=5):
        draw = ImageDraw.Draw(image)
        for _ in range(count):
            x1 = random.randint(0, self.width)
            y1 = random.randint(0, self.height)
            x2 = random.randint(0, self.width)
            y2 = random.randint(0, self.height)
            draw.line([(x1, y1), (x2, y2)], fill=(random.randint(50, 150), random.randint(50, 150), random.randint(50, 150)), width=random.randint(1, 2))
        return image
    
    def draw_text_with_transform(self, image, text):
        draw = ImageDraw.Draw(image)
        chars = list(text)
        total_width = sum(self.get_font_size(c)[0] for c in chars)
        offset_x = (self.width - total_width) // 2
        _, font_height = self.get_font_size(text)
        offset_y = (self.height - font_height) // 2
        
        for i, char in enumerate(chars):
            char_width, _ = self.get_font_size(char)
            char_image = Image.new('RGBA', (char_width + 10, self.height), (0, 0, 0, 0))
            char_draw = ImageDraw.Draw(char_image)
            
            color = (random.randint(20, 120), random.randint(20, 120), random.randint(20, 120))
            char_draw.text((5, (self.height - font_height) // 2), char, font=self.font, fill=color)
            
            angle = random.randint(-15, 15)
            char_image = char_image.rotate(angle, expand=1)
            
            pos_x = offset_x + sum(self.get_font_size(chars[j])[0] for j in range(i))
            pos_y = offset_y + random.randint(-5, 5)
            
            image.paste(char_image, (pos_x, pos_y), char_image)
        
        return image
    
    def generate_basic_captcha(self):
        image = Image.new('RGB', (self.width, self.height), (255, 255, 255))
        text = self.generate_random_text()
        draw = ImageDraw.Draw(image)
        text_width, text_height = self.get_font_size(text)
        draw.text(((self.width - text_width) // 2, 
                   (self.height - text_height) // 2), 
                  text, font=self.font, fill=(0, 0, 0))
        return image, text
    
    def generate_noisy_captcha(self):
        image = self.create_noisy_background()
        text = self.generate_random_text()
        image = self.draw_text_with_transform(image, text)
        image = self.add_noise_lines(image, random.randint(3, 6))
        image = self.add_noise_dots(image, random.randint(50, 150))
        return image, text
    
    def generate_gradient_captcha(self):
        image = self.create_gradient_background()
        text = self.generate_random_text()
        image = self.draw_text_with_transform(image, text)
        image = self.add_noise_lines(image, random.randint(2, 4))
        return image, text
    
    def generate_checkerboard_captcha(self):
        image = self.create_checkerboard_background()
        text = self.generate_random_text(char_set=CHARS_DIGITS)
        image = self.draw_text_with_transform(image, text)
        image = self.add_noise_dots(image, random.randint(30, 80))
        return image, text
    
    def generate_chinese_captcha(self):
        image = Image.new('RGB', (self.width, self.height), (255, 255, 255))
        text = self.generate_random_text(length=4, char_set=CHARS_CHINESE)
        try:
            font = ImageFont.truetype('simhei.ttf', 32)
        except:
            font = self.font
        
        draw = ImageDraw.Draw(image)
        text_width, text_height = self.get_font_size(text)
        draw.text(((self.width - text_width) // 2, 
                   (self.height - text_height) // 2), 
                  text, font=font, fill=(0, 0, 0))
        
        image = self.add_noise_lines(image, random.randint(3, 5))
        image = self.add_noise_dots(image, random.randint(50, 100))
        return image, text
    
    def generate_math_captcha(self):
        image = Image.new('RGB', (self.width, self.height), (255, 255, 255))
        num1 = random.randint(1, 99)
        num2 = random.randint(1, 99)
        operator = random.choice(['+', '-', '*'])
        
        if operator == '+':
            answer = num1 + num2
        elif operator == '-':
            answer = num1 - num2
            if answer < 0:
                num1, num2 = num2, num1
                answer = num1 - num2
        else:
            num1 = random.randint(1, 9)
            num2 = random.randint(1, 9)
            answer = num1 * num2
        
        text = f"{num1}{operator}{num2}=?"
        draw = ImageDraw.Draw(image)
        text_width, text_height = self.get_font_size(text)
        draw.text(((self.width - text_width) // 2, 
                   (self.height - text_height) // 2), 
                  text, font=self.font, fill=(0, 0, 0))
        
        image = self.add_noise_lines(image, random.randint(2, 4))
        return image, str(answer)
    
    def generate_blur_captcha(self):
        image = Image.new('RGB', (self.width, self.height), (255, 255, 255))
        text = self.generate_random_text(length=4, char_set=CHARS_DIGITS)
        draw = ImageDraw.Draw(image)
        text_width, text_height = self.get_font_size(text)
        draw.text(((self.width - text_width) // 2, 
                   (self.height - text_height) // 2), 
                  text, font=self.font, fill=(0, 0, 0))
        
        image = image.filter(ImageFilter.BLUR)
        image = self.add_noise_dots(image, random.randint(30, 60))
        return image, text
    
    def generate_colorful_captcha(self):
        image = Image.new('RGB', (self.width, self.height), (255, 255, 255))
        text = self.generate_random_text()
        draw = ImageDraw.Draw(image)
        
        total_width = sum(self.get_font_size(c)[0] for c in text)
        offset_x = (self.width - total_width) // 2
        _, font_height = self.get_font_size(text)
        offset_y = (self.height - font_height) // 2
        
        current_x = offset_x
        for char in text:
            char_width = self.get_font_size(char)[0]
            color = (random.randint(0, 200), random.randint(0, 200), random.randint(0, 200))
            draw.text((current_x, offset_y), char, font=self.font, fill=color)
            current_x += char_width
        
        image = self.add_noise_lines(image, random.randint(3, 5))
        return image, text
    
    def generate(self, captcha_type='basic'):
        methods = {
            'basic': self.generate_basic_captcha,
            'noisy': self.generate_noisy_captcha,
            'gradient': self.generate_gradient_captcha,
            'checkerboard': self.generate_checkerboard_captcha,
            'chinese': self.generate_chinese_captcha,
            'math': self.generate_math_captcha,
            'blur': self.generate_blur_captcha,
            'colorful': self.generate_colorful_captcha
        }
        
        if captcha_type in methods:
            return methods[captcha_type]()
        else:
            raise ValueError(f"Unknown captcha type: {captcha_type}")
    
    @staticmethod
    def get_available_types():
        return ['basic', 'noisy', 'gradient', 'checkerboard', 'chinese', 'math', 'blur', 'colorful']