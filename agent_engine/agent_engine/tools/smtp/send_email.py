"""SMTP Send Email — send emails via SMTP on behalf of the user."""

import asyncio
import logging
import re
import smtplib
from email.mime.multipart import MIMEMultipart
from email.mime.text import MIMEText

from ..base import BuiltinTool

logger = logging.getLogger(__name__)

_EMAIL_RE = re.compile(r"^[^@\s]+@[^@\s]+\.[^@\s]+$")
_MAX_RECIPIENTS = 10

SMTP_SEND_EMAIL_PARAMETERS = {
    "type": "object",
    "properties": {
        "to": {
            "type": "string",
            "description": "Recipient email address(es), separated by commas for multiple recipients",
        },
        "subject": {
            "type": "string",
            "description": "Email subject line",
        },
        "body": {
            "type": "string",
            "description": "Email body content (plain text or HTML depending on content_type)",
        },
        "cc": {
            "type": "string",
            "description": "CC recipient(s), separated by commas",
        },
        "bcc": {
            "type": "string",
            "description": "BCC recipient(s), separated by commas",
        },
        "reply_to": {
            "type": "string",
            "description": "Reply-To email address",
        },
        "content_type": {
            "type": "string",
            "enum": ["text", "html"],
            "description": "Content type of the email body: 'text' for plain text (default) or 'html' for HTML",
            "default": "text",
        },
    },
    "required": ["to", "subject", "body"],
}


def _parse_emails(value: str) -> list[str]:
    """Split a comma-separated string of emails and strip whitespace."""
    if not value:
        return []
    return [e.strip() for e in value.split(",") if e.strip()]


def _validate_email(email: str) -> bool:
    """Basic email format validation."""
    return bool(_EMAIL_RE.match(email))


class SmtpSendEmailTool(BuiltinTool):
    """Send emails via SMTP.

    Composes and sends emails on behalf of the user through a configured
    SMTP server. Supports plain text and HTML content, multiple recipients,
    CC, BCC, and Reply-To headers.
    """

    name = "smtp_send_email"
    description = (
        "Send an email via SMTP. Use this to compose and send emails on behalf "
        "of the user. Supports plain text and HTML content, multiple recipients "
        "(TO, CC, BCC), and Reply-To headers."
    )
    parameters = SMTP_SEND_EMAIL_PARAMETERS
    version = "1.0.0"
    idempotent = False
    side_effect_level = "write"

    @staticmethod
    def _resolve_smtp_config(context: dict | None) -> dict | None:
        """Resolve SMTP configuration from context auth_config.

        The auth_config is provided by the tool configuration set up in the
        frontend dashboard. Expected fields: smtp_host, smtp_port, smtp_user,
        smtp_password, from_address.
        """
        if not context:
            return None

        auth_config = context.get("auth_config", {})
        if not isinstance(auth_config, dict):
            return None

        host = auth_config.get("smtp_host", "")
        if not host:
            return None

        port = 587
        raw_port = auth_config.get("smtp_port", "")
        if raw_port:
            try:
                port = int(raw_port)
            except (ValueError, TypeError):
                pass

        user = auth_config.get("smtp_user", "")
        password = auth_config.get("smtp_password", "")
        from_address = auth_config.get("from_address", "")

        return {
            "host": host,
            "port": port,
            "user": user,
            "password": password,
            "from_address": from_address or user,
        }

    async def execute(self, arguments: dict, context: dict | None = None) -> dict:
        # --- Validate required fields ---
        to_raw = arguments.get("to", "").strip()
        if not to_raw:
            return {
                "status": "error",
                "data": None,
                "error": "to parameter is required",
            }

        subject = arguments.get("subject", "").strip()
        if not subject:
            return {
                "status": "error",
                "data": None,
                "error": "subject parameter is required",
            }

        body = arguments.get("body", "").strip()
        if not body:
            return {
                "status": "error",
                "data": None,
                "error": "body parameter is required",
            }

        # --- Parse and validate emails ---
        to_list = _parse_emails(to_raw)
        cc_list = _parse_emails(arguments.get("cc", ""))
        bcc_list = _parse_emails(arguments.get("bcc", ""))
        reply_to = arguments.get("reply_to", "").strip()

        all_recipients = to_list + cc_list + bcc_list
        if not all_recipients:
            return {
                "status": "error",
                "data": None,
                "error": "at least one valid recipient is required",
            }

        if len(all_recipients) > _MAX_RECIPIENTS:
            return {
                "status": "error",
                "data": None,
                "error": f"too many recipients ({len(all_recipients)}), maximum is {_MAX_RECIPIENTS}",
            }

        for email in all_recipients:
            if not _validate_email(email):
                return {
                    "status": "error",
                    "data": None,
                    "error": f"invalid email format: {email}",
                }

        if reply_to and not _validate_email(reply_to):
            return {
                "status": "error",
                "data": None,
                "error": f"invalid reply_to email format: {reply_to}",
            }

        # --- Resolve SMTP config ---
        smtp_config = self._resolve_smtp_config(context)
        if not smtp_config:
            return {
                "status": "error",
                "data": None,
                "error": "SMTP not configured (configure SMTP settings in the tool's auth config)",
            }

        from_address = smtp_config["from_address"]
        if from_address and not _validate_email(from_address):
            return {
                "status": "error",
                "data": None,
                "error": f"invalid from_address email format: {from_address}",
            }

        # --- Build email message ---
        content_type = arguments.get("content_type", "text")
        if content_type not in ("text", "html"):
            content_type = "text"

        msg = MIMEMultipart()
        msg["From"] = from_address
        msg["To"] = ", ".join(to_list)
        msg["Subject"] = subject

        if cc_list:
            msg["Cc"] = ", ".join(cc_list)
        if reply_to:
            msg["Reply-To"] = reply_to

        subtype = "html" if content_type == "html" else "plain"
        msg.attach(MIMEText(body, subtype, "utf-8"))

        # --- Send via SMTP in a thread ---
        try:
            message_id = await asyncio.to_thread(
                self._send_smtp, smtp_config, msg, all_recipients
            )
        except smtplib.SMTPException as e:
            logger.error("SMTP send failed: %s", e)
            return {
                "status": "error",
                "data": None,
                "error": f"SMTP error: {e}",
            }
        except OSError as e:
            logger.error("SMTP connection failed: %s", e)
            return {
                "status": "error",
                "data": None,
                "error": f"SMTP connection error: {e}",
            }

        return {
            "status": "success",
            "data": {
                "to": ", ".join(to_list),
                "subject": subject,
                "message_id": message_id,
            },
            "error": None,
        }

    @staticmethod
    def _send_smtp(config: dict, msg: MIMEMultipart, recipients: list[str]) -> str:
        """Send the email synchronously (called via asyncio.to_thread)."""
        host = config["host"]
        port = config["port"]
        user = config["user"]
        password = config["password"]

        if port == 465:
            server = smtplib.SMTP_SSL(host, port, timeout=30)
        else:
            server = smtplib.SMTP(host, port, timeout=30)
            server.starttls()

        try:
            if user and password:
                server.login(user, password)
            server.send_message(msg, to_addrs=recipients)
            return msg.get("Message-ID", "")
        finally:
            server.quit()
