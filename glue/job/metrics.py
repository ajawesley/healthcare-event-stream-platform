# glue/job/metrics.py
import os
from typing import Dict

import boto3
from botocore.exceptions import BotoCoreError, ClientError

from hesp_logging import get_logger

log = get_logger()


class Metrics:
    """
    Simple in-memory counter accumulator for Glue jobs.
    """

    def __init__(self):
        self.counters: Dict[str, float] = {}

    def increment(self, key: str, value: int = 1) -> None:
        self.counters[key] = self.counters.get(key, 0) + value

    def _emit_cloudwatch(self) -> None:
        if not self.counters:
            return

        namespace = os.getenv("HESP_METRICS_NAMESPACE", "HESP/Glue")
        dims = [
            {"Name": "service", "Value": "hesp-glue"},
        ]

        metric_data = []
        for name, value in self.counters.items():
            metric_data.append(
                {
                    "MetricName": name,
                    "Dimensions": dims,
                    "Value": float(value),
                    "Unit": "Count",
                }
            )

        try:
            client = boto3.client("cloudwatch")
            client.put_metric_data(Namespace=namespace, MetricData=metric_data)
            log.info("cloudwatch_metrics_emitted", namespace=namespace, metric_count=len(metric_data))
        except (BotoCoreError, ClientError) as e:
            log.error(
                "cloudwatch_metrics_failed",
                error=e,
                error_code="cloudwatch_put_metric_error",
                error_reason="failed to emit metrics to CloudWatch",
            )

    def log(self) -> None:
        log.info("glue_metrics_start")

        for key, value in self.counters.items():
            log.info(
                "glue_metric",
                metric_name=key,
                metric_value=value,
            )

        self._emit_cloudwatch()
        log.info("glue_metrics_end")
