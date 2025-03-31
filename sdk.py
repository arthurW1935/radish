import socket

class CacheClient:
    def __init__(self, hostname="localhost"):
        self.hostname = hostname
        self.port = 7171

    def _send_command(self, command: str) -> str:
        """Sends a RESP command and receives the response."""
        try:
            with socket.create_connection((self.hostname, self.port)) as client:
                client.sendall(command.encode())
                response = client.recv(1024).decode()
                return response.strip()
        except Exception as e:
            return f"ERROR: {e}"

    def put(self, key: str, value: str) -> bool:
        """PUT method for setting key-value pairs in the cache."""
        command = f"*3\r\n$3\r\nPUT\r\n${len(key)}\r\n{key}\r\n${len(value)}\r\n{value}\r\n"
        response = self._send_command(command)
        return response == "+OK"

    def get(self, key: str) -> str:
        """GET method for retrieving values from the cache."""
        command = f"*2\r\n$3\r\nGET\r\n${len(key)}\r\n{key}\r\n"
        response = self._send_command(command)
        
        if response.startswith("$-1"):  # Redis-style nil response
            return None
        
        return response.split("\r\n", 1)[-1]  # Extract the value

# Example usage
if __name__ == "__main__":
    client = CacheClient("localhost")  # Change this to the actual host
    client.put("test_key", "test_value")
    print(client.get("test_key"))
