"""Tests for the SMTP Send Email tool."""

import asyncio
import smtplib
import unittest
from unittest.mock import patch

from .send_email import SMTP_SEND_EMAIL_PARAMETERS, SmtpSendEmailTool


class TestSmtpSendEmailTool(unittest.TestCase):
    def setUp(self):
        self.tool = SmtpSendEmailTool()
        self.smtp_context = {
            "auth_config": {
                "smtp_host": "smtp.example.com",
                "smtp_port": "587",
                "smtp_user": "user@example.com",
                "smtp_password": "secret",
                "from_address": "sender@example.com",
            }
        }

    def _run(self, coro):
        return asyncio.get_event_loop().run_until_complete(coro)

    def test_name_and_description(self):
        self.assertEqual(self.tool.name, "smtp_send_email")
        self.assertIn("email", self.tool.description.lower())
        self.assertIn("SMTP", self.tool.description)
        self.assertFalse(self.tool.idempotent)
        self.assertEqual(self.tool.side_effect_level, "write")
        self.assertEqual(self.tool.version, "1.0.0")

    def test_parameters_schema(self):
        self.assertEqual(SMTP_SEND_EMAIL_PARAMETERS["type"], "object")
        props = SMTP_SEND_EMAIL_PARAMETERS["properties"]
        self.assertIn("to", props)
        self.assertIn("subject", props)
        self.assertIn("body", props)
        self.assertIn("cc", props)
        self.assertIn("bcc", props)
        self.assertIn("reply_to", props)
        self.assertIn("content_type", props)
        self.assertEqual(SMTP_SEND_EMAIL_PARAMETERS["required"], ["to", "subject", "body"])
        self.assertEqual(props["to"]["type"], "string")
        self.assertEqual(props["content_type"]["enum"], ["text", "html"])

    def test_missing_to(self):
        result = self._run(
            self.tool.execute({"subject": "Hi", "body": "Hello"}, context=self.smtp_context)
        )
        self.assertEqual(result["status"], "error")
        self.assertIn("to", result["error"])

    def test_missing_subject(self):
        result = self._run(
            self.tool.execute({"to": "a@b.com", "body": "Hello"}, context=self.smtp_context)
        )
        self.assertEqual(result["status"], "error")
        self.assertIn("subject", result["error"])

    def test_missing_body(self):
        result = self._run(
            self.tool.execute({"to": "a@b.com", "subject": "Hi"}, context=self.smtp_context)
        )
        self.assertEqual(result["status"], "error")
        self.assertIn("body", result["error"])

    def test_missing_smtp_config(self):
        result = self._run(
            self.tool.execute({"to": "a@b.com", "subject": "Hi", "body": "Hello"})
        )
        self.assertEqual(result["status"], "error")
        self.assertIn("SMTP not configured", result["error"])

    def test_missing_smtp_config_empty_auth(self):
        result = self._run(
            self.tool.execute(
                {"to": "a@b.com", "subject": "Hi", "body": "Hello"},
                context={"auth_config": {}},
            )
        )
        self.assertEqual(result["status"], "error")
        self.assertIn("SMTP not configured", result["error"])

    def test_invalid_email_format(self):
        result = self._run(
            self.tool.execute(
                {"to": "not-an-email", "subject": "Hi", "body": "Hello"},
                context=self.smtp_context,
            )
        )
        self.assertEqual(result["status"], "error")
        self.assertIn("invalid email format", result["error"])

    @patch("agent_engine.tools.smtp.send_email.SmtpSendEmailTool._send_smtp")
    def test_successful_send(self, mock_send):
        mock_send.return_value = "<msg-id-123@example.com>"

        result = self._run(
            self.tool.execute(
                {
                    "to": "recipient@example.com",
                    "subject": "Test Subject",
                    "body": "Hello, World!",
                },
                context=self.smtp_context,
            )
        )

        self.assertEqual(result["status"], "success")
        self.assertIsNone(result["error"])
        self.assertEqual(result["data"]["to"], "recipient@example.com")
        self.assertEqual(result["data"]["subject"], "Test Subject")
        self.assertEqual(result["data"]["message_id"], "<msg-id-123@example.com>")

        mock_send.assert_called_once()
        config = mock_send.call_args[0][0]
        self.assertEqual(config["host"], "smtp.example.com")
        self.assertEqual(config["port"], 587)

    @patch("agent_engine.tools.smtp.send_email.SmtpSendEmailTool._send_smtp")
    def test_successful_html_email(self, mock_send):
        mock_send.return_value = "<html-msg@example.com>"

        result = self._run(
            self.tool.execute(
                {
                    "to": "recipient@example.com",
                    "subject": "HTML Test",
                    "body": "<h1>Hello</h1><p>World</p>",
                    "content_type": "html",
                },
                context=self.smtp_context,
            )
        )

        self.assertEqual(result["status"], "success")
        mock_send.assert_called_once()
        msg = mock_send.call_args[0][1]
        payload = msg.get_payload(0)
        self.assertEqual(payload.get_content_subtype(), "html")

    @patch("agent_engine.tools.smtp.send_email.SmtpSendEmailTool._send_smtp")
    def test_smtp_error(self, mock_send):
        mock_send.side_effect = smtplib.SMTPException("Connection refused")

        result = self._run(
            self.tool.execute(
                {
                    "to": "recipient@example.com",
                    "subject": "Test",
                    "body": "Hello",
                },
                context=self.smtp_context,
            )
        )

        self.assertEqual(result["status"], "error")
        self.assertIn("SMTP error", result["error"])
        self.assertIn("Connection refused", result["error"])

    def test_too_many_recipients(self):
        emails = ", ".join([f"user{i}@example.com" for i in range(11)])
        result = self._run(
            self.tool.execute(
                {"to": emails, "subject": "Test", "body": "Hello"},
                context=self.smtp_context,
            )
        )
        self.assertEqual(result["status"], "error")
        self.assertIn("too many recipients", result["error"])

    def test_smtp_config_from_auth_config(self):
        context = {
            "auth_config": {
                "smtp_host": "mail.corp.com",
                "smtp_port": "465",
                "smtp_user": "corp@corp.com",
                "smtp_password": "corp-secret",
                "from_address": "noreply@corp.com",
            }
        }
        config = self.tool._resolve_smtp_config(context)

        self.assertIsNotNone(config)
        self.assertEqual(config["host"], "mail.corp.com")
        self.assertEqual(config["port"], 465)
        self.assertEqual(config["user"], "corp@corp.com")
        self.assertEqual(config["password"], "corp-secret")
        self.assertEqual(config["from_address"], "noreply@corp.com")

    def test_smtp_config_none_context(self):
        config = self.tool._resolve_smtp_config(None)
        self.assertIsNone(config)

    def test_smtp_config_from_address_defaults_to_user(self):
        context = {
            "auth_config": {
                "smtp_host": "smtp.test.com",
                "smtp_user": "me@test.com",
                "smtp_password": "pw",
            }
        }
        config = self.tool._resolve_smtp_config(context)
        self.assertIsNotNone(config)
        self.assertEqual(config["from_address"], "me@test.com")

    def test_registered_in_builtins(self):
        from ..registry import get_tool as get_builtin_tool
        from ..registry import is_builtin as is_builtin_tool

        self.assertTrue(is_builtin_tool("smtp_send_email"))
        tool = get_builtin_tool("smtp_send_email")
        self.assertIsNotNone(tool)
        self.assertIsInstance(tool, SmtpSendEmailTool)

    @patch("agent_engine.tools.smtp.send_email.SmtpSendEmailTool._send_smtp")
    def test_cc_and_bcc_recipients(self, mock_send):
        mock_send.return_value = "<cc-msg@example.com>"

        result = self._run(
            self.tool.execute(
                {
                    "to": "main@example.com",
                    "subject": "CC Test",
                    "body": "Hello",
                    "cc": "cc1@example.com, cc2@example.com",
                    "bcc": "bcc@example.com",
                },
                context=self.smtp_context,
            )
        )

        self.assertEqual(result["status"], "success")
        recipients = mock_send.call_args[0][2]
        self.assertIn("main@example.com", recipients)
        self.assertIn("cc1@example.com", recipients)
        self.assertIn("cc2@example.com", recipients)
        self.assertIn("bcc@example.com", recipients)

    def test_invalid_reply_to(self):
        result = self._run(
            self.tool.execute(
                {
                    "to": "a@b.com",
                    "subject": "Hi",
                    "body": "Hello",
                    "reply_to": "bad-address",
                },
                context=self.smtp_context,
            )
        )
        self.assertEqual(result["status"], "error")
        self.assertIn("reply_to", result["error"])


if __name__ == "__main__":
    unittest.main()
