"""SMTP email sending tool package."""

from .send_email import SMTP_SEND_EMAIL_PARAMETERS, SmtpSendEmailTool

__all__ = [
    "SmtpSendEmailTool",
    "SMTP_SEND_EMAIL_PARAMETERS",
]
