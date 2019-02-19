import json
import requests
import os
import unittest

import datadiff

BASE_URL = os.getenv('API_URL', 'http://localhost:8000')


class MarketsTest(unittest.TestCase):

    def test_missing_from(self):
        response = requests.get(f'{BASE_URL}/markets/ETH-ADA', params = {
            'to':   '2019-02-19T11:01:21Z',
        })
        self.assertEqual(response.status_code, 400)

    def test_missing_to(self):
        response = requests.get(f'{BASE_URL}/markets/ETH-ADA', params = {
            'from': '2019-02-19T10:59:00Z',
        })
        self.assertEqual(response.status_code, 400)

    def test_get_valid_market_info(self):
        response = requests.get(f'{BASE_URL}/markets/ETH-ADA', params = {
            'from': '2019-02-19T10:59:00Z',
            'to':   '2019-02-19T11:01:21Z',
        })
        self.assertEqual(response.status_code, 200)
        try:
            payload = response.json()
        except ValueError:
            self.assertTrue(False, 'failed to decode the payload')

        with open('tests/data/query-results.json') as f: expected = json.load(f)

        diff = datadiff.diff(payload, expected).stringify()

        self.assertEqual(len(diff), 0, 'outcome different than expected, diff:\n%s' %diff)
