import requests
import os
import unittest

BASE_URL = os.getenv('API_URL', 'http://localhost:8000')


class MarketsTest(unittest.TestCase):

    def test_get_existing_market(self):
        response = requests.get(f'{BASE_URL}/markets/ETH-ADA')
        self.assertEqual(response.status_code, 200)
        try:
            payload = response.json()
        except ValueError:
            self.assertTrue(False, 'failed to decode the payload')
        self.assertIn('low', payload)
        self.assertIs(type(payload['low']), float)
        self.assertIn('high', payload)
        self.assertIs(type(payload['high']), float)
        self.assertIn('volume', payload)
        self.assertIs(type(payload['volume']), float)

    def test_get_nonexisting_market(self):
        response = requests.get(f'{BASE_URL}/markets/invalid-223412fa')
        self.assertEqual(response.status_code, 404)
