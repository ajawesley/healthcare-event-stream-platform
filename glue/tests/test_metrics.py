# tests/test_metrics.py
from metrics import Metrics


def test_metrics_increment_and_log(caplog):
    m = Metrics()
    m.increment("records_total", 5)
    m.increment("records_total", 3)
    m.increment("records_failed", 1)

    m.log()

    assert m.counters["records_total"] == 8
    assert m.counters["records_failed"] == 1
