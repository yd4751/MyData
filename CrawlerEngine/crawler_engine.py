# crawler_engine.py

import requests
import requests.exceptions
import logging
import time
import os
from concurrent.futures import ThreadPoolExecutor, as_completed

class CrawlerEngine:
    def __init__(self, max_workers=5, default_headers=None, default_cookies=None, max_retries=3, request_interval=1.0):
        self.max_workers = max_workers
        self.tasks = []
        self.default_headers = default_headers or {}
        self.default_cookies = default_cookies or {}
        self.max_retries = max_retries
        self.request_interval = request_interval
        self.executor = ThreadPoolExecutor(max_workers=self.max_workers)
        self.last_request_time = 0
        # Ensure logs directory exists
        log_dir = 'logs'
        os.makedirs(log_dir, exist_ok=True)
        
        # Configure logging to both console and file
        log_file = os.path.join(log_dir, 'log.log')
        logging.basicConfig(
            level=logging.INFO,
            format='%(asctime)s - %(levelname)s - %(message)s',
            handlers=[
                logging.FileHandler(log_file),
                logging.StreamHandler()
            ]
        )
        self.logger = logging.getLogger(__name__)
        
    def create_task(self, url, method='GET'):
        """创建新任务辅助方法"""
        return TaskBuilder(self, url, method)

    def add_task(self, task, priority=0, callback=None):
        """添加任务到队列"""
        # Use callback from task if available, otherwise use provided callback
        task_callback = task.get('callback', callback)
        self.tasks.append((priority, task, task_callback))
        self.logger.info(f"Added task: {task.get('url')}")
    
    def run(self):
        """运行所有任务"""
        while self.tasks:
            # Sort tasks by priority (higher priority first)
            self.tasks.sort(key=lambda x: -x[0])
            
            current_tasks = self.tasks.copy()
            self.tasks = []
            
            futures = []
            for _, task, callback in current_tasks:
                future = self.executor.submit(self._execute_task, task, callback)
                futures.append(future)
            
            for future in as_completed(futures):
                try:
                    future.result()
                except Exception as e:
                    self.logger.error(f"Task failed: {e}")
    
    def _execute_task(self, task, callback):
        url = task.get('url')
        method = task.get('method', 'GET').upper()
        task_headers = task.get('headers', {})
        task_cookies = task.get('cookies', {})
        optional = task.get('optional', {})
        body = optional.pop('body', None)
        
        # Merge headers with default_headers (task headers take precedence)
        headers = {**self.default_headers, **task_headers}
        # Merge cookies with default_cookies (task cookies take precedence)
        cookies = {**self.default_cookies, **task_cookies}
        
        self.logger.info(f"Starting task: {method} {url}")
        
        # Rate limiting
        current_time = time.time()
        elapsed = current_time - self.last_request_time
        if elapsed < self.request_interval:
            time.sleep(self.request_interval - elapsed)
        
        retry_count = 0
        while retry_count <= self.max_retries:
            try:
                self.last_request_time = time.time()
                
                if method == 'GET':
                    response = requests.get(url, headers=headers, cookies=cookies, **optional)
                elif method == 'POST':
                    response = requests.post(url, headers=headers, cookies=cookies, data=body, **optional)
                else:
                    raise ValueError(f"Unsupported HTTP method: {method}")
                
                self.logger.info(f"Task completed: {method} {url} - Status: {response.status_code}")
                
                # 确保回调函数存在
                if callback:
                    custom = task.get('custom')
                    result = callback(self, {
                        'status_code': response.status_code,
                        'content': response.content,
                        'url': url
                    }, custom)
                    if result and isinstance(result, list):
                        for new_task in result:
                            self.add_task(new_task)
                break
                
            except Exception as e:
                if isinstance(e, requests.exceptions.RequestException):
                    retry_count += 1
                    if retry_count > self.max_retries:
                        error_msg = f"Task failed after {self.max_retries} retries: {method} {url} - Error: {str(e)}"
                        self.logger.error(error_msg)
                        if callback:
                            callback(self, {'error': str(e), 'url': url}, task.get('custom'))
                    else:
                        self.logger.warning(f"Retry {retry_count}/{self.max_retries}: {method} {url}")
                        time.sleep(self.request_interval * retry_count)  # Exponential backoff
                else:
                    error_msg = f"Task failed: {method} {url} - Error: {str(e)}"
                    self.logger.error(error_msg)
                    if callback:
                        callback(self, {'error': str(e), 'url': url}, task.get('custom'))
                    break



class TaskBuilder:
    """任务构造器，支持链式调用"""
    def __init__(self, engine, url, method='GET'):
        self.engine = engine
        self.task = {
            'url': url,
            'method': method.upper(),
            'headers': {},
            'cookies': {},
            'optional': {}
        }
        
    def with_headers(self, headers):
        """设置请求头"""
        self.task['headers'].update(headers)
        return self
        
    def with_cookies(self, cookies):
        """设置cookies"""
        self.task['cookies'].update(cookies)
        return self
        
    def with_body(self, body):
        """设置请求体"""
        self.task['optional']['body'] = body
        return self
        
    def with_timeout(self, timeout):
        """设置超时时间"""
        self.task['optional']['timeout'] = timeout
        return self
        
    def with_callback(self, callback):
        """设置回调函数"""
        self.task['callback'] = callback
        return self
        
    def with_custom(self, custom):
        """设置自定义上下文"""
        self.task['custom'] = custom
        return self
        
    def build(self):
        """构建最终任务字典"""
        return self.task