from collections import defaultdict


class Graph:
    def __init__(self, directed=False):
        self.adj_list = defaultdict(list)
        self.directed = directed

    def add_vertex(self, vertex):
        if vertex not in self.adj_list:
            self.adj_list[vertex] = []

    def add_edge(self, vertex1, vertex2, weight=None):
        self.add_vertex(vertex1)
        self.add_vertex(vertex2)

        edge_data = {"to": vertex2, "weight": weight} if weight is not None else {"to": vertex2}
        self.adj_list[vertex1].append(edge_data)

        if not self.directed:
            edge_data_reverse = {"to": vertex1, "weight": weight} if weight is not None else {"to": vertex1}
            self.adj_list[vertex2].append(edge_data_reverse)

    def get_neighbors(self, vertex):
        return [edge["to"] for edge in self.adj_list.get(vertex, [])]

    def get_edges(self, vertex):
        return self.adj_list.get(vertex, [])

    def has_vertex(self, vertex):
        return vertex in self.adj_list

    def has_edge(self, vertex1, vertex2):
        for edge in self.adj_list.get(vertex1, []):
            if edge["to"] == vertex2:
                return True
        return False

    def get_vertices(self):
        return list(self.adj_list.keys())

    def get_edge_count(self):
        count = sum(len(edges) for edges in self.adj_list.values())
        return count // 2 if not self.directed else count

    def to_dict(self):
        return dict(self.adj_list)


def handler(event):
    directed = event.get("directed", False)
    edges = event.get("edges", [])
    operations = event.get("operations", [])

    if not isinstance(edges, list):
        return {"ok": False, "error": "edges must be an array"}

    if not isinstance(operations, list):
        return {"ok": False, "error": "operations must be an array"}

    try:
        graph = Graph(directed)

        # Add initial edges
        for edge in edges:
            if not isinstance(edge, dict):
                return {"ok": False, "error": "each edge must be an object"}

            vertex1 = edge.get("from") or edge.get("vertex1")
            vertex2 = edge.get("to") or edge.get("vertex2")
            weight = edge.get("weight")

            if vertex1 is None or vertex2 is None:
                return {"ok": False, "error": "edges must have 'from'/'to' or 'vertex1'/'vertex2' fields"}

            graph.add_edge(vertex1, vertex2, weight)

        results = []
        for op in operations:
            if not isinstance(op, dict):
                return {"ok": False, "error": "each operation must be an object"}

            action = op.get("action")
            if action not in ["add_vertex", "add_edge", "get_neighbors", "get_edges", "has_vertex", "has_edge", "vertices", "edge_count"]:
                return {"ok": False, "error": f"unknown action: {action}"}

            if action == "add_vertex":
                vertex = op.get("vertex")
                if vertex is None:
                    return {"ok": False, "error": "add_vertex action requires 'vertex'"}
                graph.add_vertex(vertex)
                results.append({"action": "add_vertex", "vertex": vertex, "success": True})

            elif action == "add_edge":
                vertex1 = op.get("from") or op.get("vertex1")
                vertex2 = op.get("to") or op.get("vertex2")
                weight = op.get("weight")
                if vertex1 is None or vertex2 is None:
                    return {"ok": False, "error": "add_edge action requires 'from'/'to' or 'vertex1'/'vertex2'"}
                graph.add_edge(vertex1, vertex2, weight)
                results.append({"action": "add_edge", "from": vertex1, "to": vertex2, "weight": weight, "success": True})

            elif action == "get_neighbors":
                vertex = op.get("vertex")
                if vertex is None:
                    return {"ok": False, "error": "get_neighbors action requires 'vertex'"}
                result = graph.get_neighbors(vertex)
                results.append({"action": "get_neighbors", "vertex": vertex, "result": result})

            elif action == "get_edges":
                vertex = op.get("vertex")
                if vertex is None:
                    return {"ok": False, "error": "get_edges action requires 'vertex'"}
                result = graph.get_edges(vertex)
                results.append({"action": "get_edges", "vertex": vertex, "result": result})

            elif action == "has_vertex":
                vertex = op.get("vertex")
                if vertex is None:
                    return {"ok": False, "error": "has_vertex action requires 'vertex'"}
                result = graph.has_vertex(vertex)
                results.append({"action": "has_vertex", "vertex": vertex, "result": result})

            elif action == "has_edge":
                vertex1 = op.get("from") or op.get("vertex1")
                vertex2 = op.get("to") or op.get("vertex2")
                if vertex1 is None or vertex2 is None:
                    return {"ok": False, "error": "has_edge action requires 'from'/'to' or 'vertex1'/'vertex2'"}
                result = graph.has_edge(vertex1, vertex2)
                results.append({"action": "has_edge", "from": vertex1, "to": vertex2, "result": result})

            elif action == "vertices":
                result = graph.get_vertices()
                results.append({"action": "vertices", "result": result})

            elif action == "edge_count":
                result = graph.get_edge_count()
                results.append({"action": "edge_count", "result": result})

        return {
            "ok": True,
            "result": {
                "adjacency_list": graph.to_dict(),
                "directed": directed,
                "operations": results
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"graph operation failed: {str(e)}"}