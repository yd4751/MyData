import torch
import torch.nn as nn
import torch.optim as optim
from torch.utils.data import DataLoader, random_split
from tqdm import tqdm
import os

from dataset import load_train_dataset, decode_label
from model import create_model
from config import MODEL_DIR, BATCH_SIZE, EPOCHS, LEARNING_RATE

def train():
    dataset = load_train_dataset()
    
    if len(dataset) == 0:
        print("Error: Training dataset is empty! Please generate data first.")
        return
    
    train_size = int(0.9 * len(dataset))
    val_size = len(dataset) - train_size
    train_dataset, val_dataset = random_split(dataset, [train_size, val_size])
    
    train_loader = DataLoader(train_dataset, batch_size=BATCH_SIZE, shuffle=True)
    val_loader = DataLoader(val_dataset, batch_size=BATCH_SIZE, shuffle=False)
    
    model = create_model()
    device = torch.device('cpu')
    print("Using CPU for training (as requested)")
    model.to(device)
    
    criterion = nn.CrossEntropyLoss()
    optimizer = optim.Adam(model.parameters(), lr=LEARNING_RATE)
    
    best_val_acc = 0.0
    
    for epoch in range(EPOCHS):
        model.train()
        train_loss = 0.0
        train_correct = 0
        train_total = 0
        
        progress_bar = tqdm(train_loader, desc=f'Epoch {epoch+1}/{EPOCHS}')
        for images, labels in progress_bar:
            images = images.to(device).float()
            labels = labels.to(device).float()
            
            optimizer.zero_grad()
            outputs = model(images)
            
            loss = 0
            for i in range(4):
                loss += criterion(outputs[:, i, :], labels[:, i, :])
            loss /= 4
            
            loss.backward()
            optimizer.step()
            
            train_loss += loss.item() * images.size(0)
            
            for i in range(images.size(0)):
                pred = decode_label(outputs[i].detach().numpy())
                actual = decode_label(labels[i].detach().numpy())
                if pred == actual:
                    train_correct += 1
                train_total += 1
            
            progress_bar.set_postfix({
                'loss': f'{train_loss/train_total:.4f}',
                'acc': f'{train_correct/train_total:.4f}'
            })
        
        train_loss /= train_total
        train_acc = train_correct / train_total
        
        model.eval()
        val_loss = 0.0
        val_correct = 0
        val_total = 0
        
        with torch.no_grad():
            for images, labels in val_loader:
                images = images.to(device).float()
                labels = labels.to(device).float()
                
                outputs = model(images)
                
                loss = 0
                for i in range(4):
                    loss += criterion(outputs[:, i, :], labels[:, i, :])
                loss /= 4
                
                val_loss += loss.item() * images.size(0)
                
                for i in range(images.size(0)):
                    pred = decode_label(outputs[i].detach().numpy())
                    actual = decode_label(labels[i].detach().numpy())
                    if pred == actual:
                        val_correct += 1
                    val_total += 1
        
        val_loss /= val_total
        val_acc = val_correct / val_total
        
        print(f'Epoch {epoch+1}/{EPOCHS}')
        print(f'Train Loss: {train_loss:.4f}, Train Acc: {train_acc:.4f}')
        print(f'Val Loss: {val_loss:.4f}, Val Acc: {val_acc:.4f}')
        print('-' * 50)
        
        if val_acc > best_val_acc:
            best_val_acc = val_acc
            model_path = os.path.join(MODEL_DIR, 'captcha_model.pth')
            torch.save(model.state_dict(), model_path)
            print(f'Model saved to {model_path}')
    
    print(f'Training complete. Best validation accuracy: {best_val_acc:.4f}')

if __name__ == '__main__':
    train()