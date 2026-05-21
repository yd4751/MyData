# test_engine.py
import unittest
from unittest.mock import patch
from crawler_engine import CrawlerEngine

class TestCrawlerEngine(unittest.TestCase):
    def setUp(self):
        self.engine = CrawlerEngine(max_workers=1)

    @patch('requests.get')
    def test_task_builder(self, mock_get):
        # 模拟响应
        mock_get.return_value.status_code = 200
        mock_get.return_value.content = b'test content'

        # 测试回调
        def callback(engine, response, custom):
            self.assertEqual(response['status_code'], 200)
            self.assertEqual(custom['test'], True)
            return None

        # 使用TaskBuilder创建任务
        task = (self.engine.create_task('http://test.com')
                .with_headers({'Test': 'true'})
                .with_custom({'test': True})
                .with_callback(callback)
                .build())

        self.engine.add_task(task=task, callback=callback)
        self.engine.run()

    @patch('requests.get')
    def test_task_chain(self, mock_get):
        # 模拟响应
        mock_get.return_value.status_code = 200
        mock_get.return_value.content = b'chain test'

        # 第一级回调
        def first_callback(engine, response, custom):
            custom = custom or {}
            custom['first'] = True
            # 创建第二级任务
            new_task = (engine.create_task('http://test.com/level2')
                       .with_callback(second_callback)
                       .with_custom(custom)
                       .build())
            return [new_task]

        # 第二级回调
        def second_callback(engine, response, custom):
            self.assertTrue(custom['first'])
            custom['second'] = True
            return None

        # 初始任务
        task = (self.engine.create_task('http://test.com')
               .with_callback(first_callback)
               .build())

        self.engine.add_task(task=task)
        self.engine.run()

if __name__ == '__main__':
    unittest.main()
