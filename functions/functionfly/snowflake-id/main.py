import time
import threading

# Snowflake ID structure:
# 1 bit: unused (always 0)
# 41 bits: timestamp in milliseconds (custom epoch)
# 5 bits: datacenter ID
# 5 bits: machine ID
# 12 bits: sequence number

EPOCH = 1609459200000  # 2021-01-01 00:00:00 UTC

class SnowflakeGenerator:
    def __init__(self, machine_id=1, datacenter_id=1):
        self.machine_id = machine_id & 0x1F
        self.datacenter_id = datacenter_id & 0x1F
        self.sequence = 0
        self.last_timestamp = -1
        self.lock = threading.Lock()

    def _timestamp(self):
        return int(time.time() * 1000) - EPOCH

    def _wait_next_millis(self, last_timestamp):
        timestamp = self._timestamp()
        while timestamp <= last_timestamp:
            timestamp = self._timestamp()
        return timestamp

    def generate(self):
        with self.lock:
            timestamp = self._timestamp()
            if timestamp == self.last_timestamp:
                self.sequence = (self.sequence + 1) & 0xFFF
                if self.sequence == 0:
                    timestamp = self._wait_next_millis(self.last_timestamp)
            else:
                self.sequence = 0
            self.last_timestamp = timestamp
            snowflake_id = ((timestamp << 22) |
                           (self.datacenter_id << 17) |
                           (self.machine_id << 12) |
                           self.sequence)
            return snowflake_id

def handler(event):
    try:
        machine_id = event.get("machine_id", 1) if isinstance(event, dict) else 1
        datacenter_id = event.get("datacenter_id", 1) if isinstance(event, dict) else 1
        generator = SnowflakeGenerator(machine_id, datacenter_id)
        snowflake_id = generator.generate()
        return {"ok": True, "id": str(snowflake_id)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
