# glue/job/hesp_logging.py
import json
import sys
import uuid
import datetime
import traceback
from typing import Any, Optional

SERVICE_NAME = "hesp-glue"
SOURCE_SYSTEM = "hesp-glue"
LOG_FORMAT = "json"


def _now_iso() -> str:
    return datetime.datetime.utcnow().replace(tzinfo=datetime.timezone.utc).isoformat()


class GlueLogger:
    def __init__(self):
        self.service = SERVICE_NAME
        self.source_system = SOURCE_SYSTEM
        self._trace_id: Optional[str] = None
        self._span_id: Optional[str] = None

    def set_trace_context(self, trace_id: Optional[str], span_id: Optional[str]) -> None:
        self._trace_id = trace_id
        self._span_id = span_id

    def _emit(self, level: str, message: str, **fields: Any):
        trace_id = fields.pop("trace_id", self._trace_id)
        span_id = fields.pop("span_id", self._span_id)

        record = {
            "level": level,
            "message": message,
            "event_id": str(uuid.uuid4()),
            "trace_id": trace_id,
            "span_id": span_id,
            "service": self.service,
            "source_system": self.source_system,
            "format": LOG_FORMAT,
            "produced_at": _now_iso(),
        }

        record.update(fields)

        sys.stdout.write(json.dumps(record) + "\n")
        sys.stdout.flush()

    def info(self, message: str, **fields: Any):
        self._emit("info", message, **fields)

    def warn(self, message: str, **fields: Any):
        self._emit("warn", message, **fields)

    def debug(self, message: str, **fields: Any):
        self._emit("debug", message, **fields)

    def error(self, message: str, error: Exception, error_code: str, error_reason: str, **fields: Any):
        tb = traceback.format_exc()
        self._emit(
            "error",
            message,
            error_code=error_code,
            error_reason=error_reason,
            error_message=str(error),
            stacktrace=tb,
            **fields,
        )


_logger = GlueLogger()


def get_logger() -> GlueLogger:
    return _logger


def log_job_started(job_name: str, run_id: str, trace_id: Optional[str] = None, span_id: Optional[str] = None):
    _logger.info(
        "glue_job_started",
        job_name=job_name,
        run_id=run_id,
        trace_id=trace_id,
        span_id=span_id,
    )


def log_job_completed(job_name: str, run_id: str, duration_ms: float, trace_id: Optional[str] = None, span_id: Optional[str] = None):
    _logger.info(
        "glue_job_completed",
        job_name=job_name,
        run_id=run_id,
        duration_ms=duration_ms,
        trace_id=trace_id,
        span_id=span_id,
    )


def log_job_failed(job_name: str, run_id: str, error: Exception, trace_id: Optional[str] = None, span_id: Optional[str] = None):
    _logger.error(
        "glue_job_failed",
        error=error,
        error_code="glue_job_failure",
        error_reason=str(error),
        job_name=job_name,
        run_id=run_id,
        trace_id=trace_id,
        span_id=span_id,
    )
