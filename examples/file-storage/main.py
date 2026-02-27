import json
import os

def handler(input_data):
    """
    Demonstrate file storage operations.
    """
    try:
        filename = os.getenv('STORAGE_FILE', 'data.json')

        if input_data.get('action') == 'write':
            # Write data to file
            data_to_store = {
                "message": input_data.get('message', 'Hello from FunctionFly'),
                "timestamp": __import__('datetime').datetime.now().isoformat(),
                "function": "file-storage"
            }

            # Convert to JSON string for storage
            json_data = json.dumps(data_to_store, indent=2)

            # In a real implementation, this would call the storage host functions
            # For now, we'll simulate the storage operation

            result = {
                "status": "success",
                "action": "write",
                "filename": filename,
                "data_size": len(json_data),
                "data": data_to_store
            }

        elif input_data.get('action') == 'read':
            # Read data from file
            # In a real implementation, this would read from storage
            # For simulation, return sample data

            sample_data = {
                "message": "Sample stored data",
                "timestamp": "2024-01-01T12:00:00",
                "function": "file-storage"
            }

            result = {
                "status": "success",
                "action": "read",
                "filename": filename,
                "data": sample_data
            }

        else:
            result = {
                "status": "error",
                "message": "Invalid action. Use 'read' or 'write'",
                "available_actions": ["read", "write"]
            }

        return result

    except Exception as e:
        return {
            "status": "error",
            "message": f"Storage operation failed: {str(e)}"
        }