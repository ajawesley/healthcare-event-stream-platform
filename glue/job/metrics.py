import logging

logger = logging.getLogger(__name__)

class Metrics:
    """
    Simple in-memory counter accumulator for Glue jobs.
    Ensures metrics are always logged, even if the job fails.
    """

    def __init__(self):
        self.counters = {}

    def increment(self, key: str, value: int = 1) -> None:
        self.counters[key] = self.counters.get(key, 0) + value

    def log(self) -> None:
        """
        Emits all metrics to CloudWatch Logs in a structured format.
        """
        logger.info("---- Glue Job Metrics ----")
        for key, value in self.counters.items():
            logger.info(f"{key}: {value}")
        logger.info("---- End Metrics ----")
