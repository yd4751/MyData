#!/usr/bin/env python3
"""
Configuration for crawler engine
"""

CRAWLER_CONFIG = {
    # Thread pool settings
    'max_workers': 10,
    
    # Request settings
    'max_retries': 3,
    'request_interval': 1.0,  # seconds between requests
    'request_timeout': 10,    # seconds
    
    # Default request headers
    'default_headers': {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36',
        'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8'
    },
    
    # Default cookies
    'default_cookies': {
        'session_id': 'default_session',
        'language': 'en-US'
    }
}
