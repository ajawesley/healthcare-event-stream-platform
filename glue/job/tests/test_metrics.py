from metrics import Metrics


def test_metrics_increment_and_log(monkeypatch):
    m = Metrics()
    m.increment("valid_records", 3)
    m.increment("valid_records", 2)
    m.increment("invalid_records", 1)

    calls = []

    from glue.job import hesp_logging

    def fake_info(msg, **fields):
        calls.append((msg, fields))

    monkeypatch.setattr(hesp_logging._logger, "info", fake_info)

    m.log()

    names = [c[0] for c in calls]
    assert "glue_metrics_start" in names
    assert "glue_metrics_end" in names
    metric_events = [f for m, f in calls if m == "glue_metric"]
    assert any(e["metric_name"] == "valid_records" and e["metric_value"] == 5 for e in metric_events)
    assert any(e["metric_name"] == "invalid_records" and e["metric_value"] == 1 for e in metric_events)
