"""
FOM Training Data Export Service

Exports FOM data for model training.
Handles train/val/test splits and streaming export.
"""

import json
from typing import Iterator, Optional


class FOMExportService:
    def __init__(self, postgres_conn, duckdb_conn=None):
        self.pg = postgres_conn
        self.duckdb = duckdb_conn

    def export_training_data(
        self,
        output_path: str,
        split: str = "train",
        limit: int = 1000000,
        goal_type: Optional[str] = None,
    ) -> int:
        query = """
            SELECT
                goal_text,
                goal_type,
                workflow_json,
                outcome_success,
                outcome_score,
                total_cost,
                total_time_ms
            FROM fom_training_records
            WHERE split = %s
            AND is_synthetic = FALSE
        """
        params = [split]

        if goal_type:
            query += " AND goal_type = %s"
            params.append(goal_type)

        query += " ORDER BY created_at DESC LIMIT %s"
        params.append(limit)

        count = 0
        with open(output_path, "w") as f:
            rows = self._execute_query(query, params)
            for row in rows:
                record = {
                    "goal": row["goal_text"],
                    "goal_type": row["goal_type"],
                    "workflow": row["workflow_json"],
                    "outcome": {
                        "success": row["outcome_success"],
                        "score": row["outcome_score"],
                        "cost": float(row["total_cost"]) if row["total_cost"] else 0.0,
                        "time_ms": row["total_time_ms"] if row["total_time_ms"] else 0,
                    },
                }
                f.write(json.dumps(record) + "\n")
                count += 1

        return count

    def export_synthetic_data(
        self,
        output_path: str,
        version: str = "synthetic_v1",
        limit: int = 500000,
    ) -> int:
        query = """
            SELECT
                goal_text,
                goal_type,
                workflow_json,
                outcome_success,
                outcome_score,
                total_cost,
                total_time_ms
            FROM fom_training_records
            WHERE is_synthetic = TRUE
            AND generation_method = %s
            ORDER BY created_at DESC
            LIMIT %s
        """

        count = 0
        with open(output_path, "w") as f:
            rows = self._execute_query(query, [version, limit])
            for row in rows:
                record = {
                    "goal": row["goal_text"],
                    "goal_type": row["goal_type"],
                    "workflow": row["workflow_json"],
                    "outcome": {
                        "success": row["outcome_success"],
                        "score": row["outcome_score"],
                        "cost": float(row["total_cost"]) if row["total_cost"] else 0.0,
                        "time_ms": row["total_time_ms"] if row["total_time_ms"] else 0,
                    },
                }
                f.write(json.dumps(record) + "\n")
                count += 1

        return count

    def export_streaming(
        self,
        split: str = "train",
        batch_size: int = 1000,
    ) -> Iterator[dict]:
        query = """
            SELECT
                id,
                goal_text,
                goal_type,
                workflow_json,
                outcome_success,
                outcome_score,
                total_cost,
                total_time_ms,
                created_at
            FROM fom_training_records
            WHERE split = %s
            AND is_synthetic = FALSE
            ORDER BY created_at DESC
        """

        offset = 0
        while True:
            batch_query = query + " LIMIT %s OFFSET %s"
            rows = self._execute_query(batch_query, [split, batch_size, offset])

            batch = list(rows)
            if not batch:
                break

            for row in batch:
                yield {
                    "id": str(row["id"]),
                    "goal": row["goal_text"],
                    "goal_type": row["goal_type"],
                    "workflow": row["workflow_json"],
                    "outcome": {
                        "success": row["outcome_success"],
                        "score": row["outcome_score"],
                        "cost": float(row["total_cost"]) if row["total_cost"] else 0.0,
                        "time_ms": row["total_time_ms"] if row["total_time_ms"] else 0,
                    },
                    "timestamp": row["created_at"].isoformat() if row["created_at"] else None,
                }

            offset += batch_size
            if len(batch) < batch_size:
                break

    def export_for_huggingface(
        self,
        output_path: str,
        split: str = "train",
        limit: int = 100000,
    ) -> int:
        prompt_template = """Goal: {goal}
Goal Type: {goal_type}

What is the best workflow to achieve this goal?"""

        response_template = """Workflow: {workflow}
Outcome: {outcome_success} (score: {outcome_score})"""

        count = 0
        with open(output_path, "w") as f:
            for record in self.export_streaming(split=split):
                if count >= limit:
                    break

                prompt = prompt_template.format(
                    goal=record["goal"],
                    goal_type=record["goal_type"],
                )
                response = response_template.format(
                    workflow=json.dumps(record["workflow"]),
                    outcome_success=record["outcome"]["success"],
                    outcome_score=record["outcome"]["score"],
                )

                formatted = {
                    "prompt": prompt,
                    "response": response,
                }
                f.write(json.dumps(formatted) + "\n")
                count += 1

        return count

    def _execute_query(self, query: str, params: list) -> list[dict]:
        if self.duckdb:
            return self._execute_duckdb(query, params)
        return self._execute_postgres(query, params)

    def _execute_postgres(self, query: str, params: list) -> list[dict]:
        import psycopg2

        conn = psycopg2.connect(
            host=self.pg.get("host", "localhost"),
            port=self.pg.get("port", 5432),
            database=self.pg.get("database", "functionfly"),
            user=self.pg.get("user", "postgres"),
            password=self.pg.get("password", ""),
        )
        try:
            with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
                cur.execute(query, params)
                return cur.fetchall()
        finally:
            conn.close()

    def _execute_duckdb(self, query: str, params: list) -> list[dict]:
        import duckdb

        conn = duckdb.connect(database=self.duckdb.get("database", ":memory:"))
        try:
            result = conn.execute(query, params).fetchall()
            return [dict(row) for row in result]
        finally:
            conn.close()


class StreamingFOMExport:
    def __init__(self, kafka_brokers: list[str], topic: str = "fom.training.stream"):
        self.kafka_brokers = kafka_brokers
        self.topic = topic
        self.producer = None

    def start(self):
        try:
            from kafka import KafkaProducer
            self.producer = KafkaProducer(
                bootstrap_servers=self.kafka_brokers,
                value_serializer=lambda v: json.dumps(v).encode("utf-8"),
            )
        except ImportError:
            print("kafka-python not installed, streaming export disabled")

    def on_fom_record(self, record: dict):
        if self.producer is None:
            self.start()

        if self.producer:
            self.producer.send(self.topic, record)

    def close(self):
        if self.producer:
            self.producer.close()
            self.producer = None