"""Tests for JSON structured logging."""

import json
import logging
import unittest

from agent_engine.main import JSONFormatter


class TestJSONFormatter(unittest.TestCase):
    def setUp(self):
        self.formatter = JSONFormatter()
        self.logger = logging.getLogger("test_json_formatter")
        self.logger.setLevel(logging.DEBUG)

    def test_produces_valid_json(self):
        record = self.logger.makeRecord(
            "test",
            logging.INFO,
            "file.py",
            1,
            "hello world",
            (),
            None,
        )
        output = self.formatter.format(record)
        data = json.loads(output)
        self.assertEqual(data["level"], "INFO")
        self.assertEqual(data["message"], "hello world")
        self.assertIn("timestamp", data)
        self.assertIn("logger", data)

    def test_extra_fields(self):
        record = self.logger.makeRecord(
            "test",
            logging.INFO,
            "file.py",
            1,
            "task done",
            (),
            None,
        )
        record.request_id = "req-123"
        record.agent_id = "agent-456"
        output = self.formatter.format(record)
        data = json.loads(output)
        self.assertEqual(data["request_id"], "req-123")
        self.assertEqual(data["agent_id"], "agent-456")

    def test_exception_included(self):
        try:
            raise ValueError("boom")
        except ValueError:
            import sys

            exc_info = sys.exc_info()

        record = self.logger.makeRecord(
            "test",
            logging.ERROR,
            "file.py",
            1,
            "error occurred",
            (),
            exc_info,
        )
        output = self.formatter.format(record)
        data = json.loads(output)
        self.assertIn("exception", data)
        self.assertIn("ValueError", data["exception"])

    def test_no_extra_fields_when_absent(self):
        record = self.logger.makeRecord(
            "test",
            logging.INFO,
            "file.py",
            1,
            "simple msg",
            (),
            None,
        )
        output = self.formatter.format(record)
        data = json.loads(output)
        self.assertNotIn("request_id", data)
        self.assertNotIn("agent_id", data)


if __name__ == "__main__":
    unittest.main()
