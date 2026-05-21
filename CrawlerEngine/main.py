# main.py

#!/usr/bin/env python3
"""
Main entry point for the crawler engine
"""

import os
import importlib.util
from crawler_engine import CrawlerEngine

def load_tasks_from_folder(folder_path):
    tasks = []
    for root, _, files in os.walk(folder_path):
        for filename in files:
            if filename.endswith('.py') and filename != '__init__.py':
                file_path = os.path.join(root, filename)
                # 使用相对路径作为模块名，避免冲突
                rel_path = os.path.relpath(file_path, folder_path)
                module_name = rel_path.replace(os.path.sep, '_')[:-3]  # 去掉.py后缀
                
                spec = importlib.util.spec_from_file_location(module_name, file_path)
                module = importlib.util.module_from_spec(spec)
                spec.loader.exec_module(module)
                
                # 假设每个任务文件中都有一个名为`task`的函数
                if hasattr(module, 'task'):
                    tasks.append(module.task)
    return tasks

def main():
    # Initialize crawler engine with custom config
    engine = CrawlerEngine(max_workers=5)
    
    # Load tasks from the specified folder
    tasks_folder = 'tasks'
    task_functions = load_tasks_from_folder(tasks_folder)
    
    for task_function in task_functions:
        task_function(engine)
    
    # Start the engine
    engine.run()

if __name__ == "__main__":
    main()
