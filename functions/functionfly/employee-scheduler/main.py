from typing import Any
from collections import defaultdict


def parse_availability(availability: dict, day: str, hour: int) -> bool:
    if not availability:
        return True
    
    if day not in availability:
        return True
    
    day_availability = availability[day]
    
    if isinstance(day_availability, bool):
        return day_availability
    
    if isinstance(day_availability, dict):
        start = day_availability.get("start", 0)
        end = day_availability.get("end", 24)
        return start <= hour < end
    
    if isinstance(day_availability, list):
        for slot in day_availability:
            if isinstance(slot, dict):
                start = slot.get("start", 0)
                end = slot.get("end", 24)
                if start <= hour < end:
                    return True
            elif isinstance(slot, (int, float)):
                if slot == hour:
                    return True
        return False
    
    return True


def is_employee_available(employee: dict, day: str, start_hour: int, end_hour: int) -> bool:
    availability = employee.get("availability", {})
    name = employee.get("name", f"Employee_{id(employee)}")
    
    for hour in range(start_hour, end_hour):
        if not parse_availability(availability, day, hour):
            return False
    
    return True


def schedule_employees(employees: list, shifts: list, min_per_shift: int) -> tuple[list, list, float]:
    schedule = []
    unassigned = []
    
    shift_assignments = defaultdict(list)
    
    for shift in shifts:
        day = shift.get("day")
        start_hour = shift.get("start_hour")
        end_hour = shift.get("end_hour")
        
        if day is None or start_hour is None or end_hour is None:
            continue
        
        if start_hour >= end_hour:
            continue
        
        shift_key = (day, start_hour, end_hour)
        
        required = min_per_shift
        assigned = 0
        
        for employee in employees:
            if assigned >= required:
                break
            
            emp_name = employee.get("name", "")
            
            if emp_name in shift_assignments[day]:
                continue
            
            if is_employee_available(employee, day, start_hour, end_hour):
                schedule.append({
                    "day": day,
                    "shift": f"{start_hour}:00-{end_hour}:00",
                    "employee": emp_name,
                    "start_hour": start_hour,
                    "end_hour": end_hour
                })
                
                shift_assignments[day].append(emp_name)
                assigned += 1
        
        if assigned < required:
            unassigned.append({
                "day": day,
                "shift": f"{start_hour}:00-{end_hour}:00",
                "start_hour": start_hour,
                "end_hour": end_hour,
                "employees_needed": required - assigned
            })
    
    total_shifts = len(shifts)
    assigned_shifts = total_shifts - len(unassigned)
    
    if total_shifts > 0:
        coverage_score = round((assigned_shifts / total_shifts) * 100, 1)
    else:
        coverage_score = 0.0
    
    return schedule, unassigned, coverage_score


def handler(event: dict[str, Any]) -> dict[str, Any]:
    try:
        employees = event.get("employees", [])
        shifts = event.get("shifts", [])
        min_per_shift = event.get("min_per_shift", 1)
        
        if not isinstance(employees, list):
            return {"ok": False, "error": "employees must be a list"}
        
        if len(employees) == 0:
            return {"ok": False, "error": "employees list cannot be empty"}
        
        if not isinstance(shifts, list):
            return {"ok": False, "error": "shifts must be a list"}
        
        if len(shifts) == 0:
            return {"ok": False, "error": "shifts list cannot be empty"}
        
        try:
            min_per_shift = int(min_per_shift)
        except (ValueError, TypeError):
            return {"ok": False, "error": "min_per_shift must be an integer"}
        
        if min_per_shift < 1:
            return {"ok": False, "error": "min_per_shift must be at least 1"}
        
        for i, emp in enumerate(employees):
            if not isinstance(emp, dict):
                return {"ok": False, "error": f"Employee at index {i} must be an object"}
            if "name" not in emp:
                return {"ok": False, "error": f"Employee at index {i} missing required field 'name'"}
        
        for i, shift in enumerate(shifts):
            if not isinstance(shift, dict):
                return {"ok": False, "error": f"Shift at index {i} must be an object"}
            if "day" not in shift:
                return {"ok": False, "error": f"Shift at index {i} missing required field 'day'"}
            if "start_hour" not in shift:
                return {"ok": False, "error": f"Shift at index {i} missing required field 'start_hour'"}
            if "end_hour" not in shift:
                return {"ok": False, "error": f"Shift at index {i} missing required field 'end_hour'"}
            
            start = shift.get("start_hour")
            end = shift.get("end_hour")
            
            if not isinstance(start, (int, float)) or not isinstance(end, (int, float)):
                return {"ok": False, "error": f"Shift at index {i} hours must be numbers"}
            
            if start < 0 or start > 24 or end < 0 or end > 24:
                return {"ok": False, "error": f"Shift at index {i} hours must be between 0 and 24"}
            
            if start >= end:
                return {"ok": False, "error": f"Shift at index {i} start_hour must be less than end_hour"}
        
        schedule, unassigned_shifts, coverage_score = schedule_employees(employees, shifts, min_per_shift)
        
        return {
            "ok": True,
            "schedule": schedule,
            "unassigned_shifts": unassigned_shifts,
            "coverage_score": coverage_score,
            "total_shifts": len(shifts),
            "assigned_count": len(schedule),
            "unassigned_count": len(unassigned_shifts)
        }
        
    except Exception as e:
        return {"ok": False, "error": str(e)}
