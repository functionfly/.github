#!/bin/bash
# Latency measurement for running services
# Auto-detects which services are available

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ITERATIONS=${1:-10}

log() { echo -e "${GREEN}[$(date +'%H:%M:%S')]${NC} $1"; }

echo ""
log "Latency Measurement Test"
log "========================"
echo "Iterations: $ITERATIONS"
echo ""

declare -A SERVICES
SERVICES=(
  ["orchestrator"]="8080"
  ["ai-service"]="18081"
  ["runtime-local"]="8787"
  ["runtime-prism"]="8084"
  ["runtime-kotlin"]="8085"
)

available=()

for svc in "${!SERVICES[@]}"; do
  port="${SERVICES[$svc]}"
  if curl -sf --max-time 1 "http://localhost:$port/health" > /dev/null 2>&1; then
    available+=("$svc:$port")
    echo -e "  ${GREEN}✓${NC} $svc (port $port)"
  else
    echo -e "  ${RED}✗${NC} $svc (port $port) - not running"
  fi
done

if [[ ${#available[@]} -eq 0 ]]; then
  echo -e "\n${RED}No services available${NC}"
  exit 1
fi

echo ""
echo "Measuring latency..."
echo ""

echo "service,url,avg_ms,p50_ms,p95_ms,min_ms,max_ms,errors" > "$SCRIPT_DIR/latency_results.csv"

for svc in "${available[@]}"; do
  name="${svc%%:*}"
  port="${svc##*:}"
  
  times=()
  errors=0
  
  for i in $(seq 1 $ITERATIONS); do
    result=$(curl -sf -w "%{time_total}" -o /dev/null "http://localhost:$port/health" 2>&1)
    
    if [[ $? -eq 0 && -n "$result" ]]; then
      ms=$(echo "$result * 1000" | bc 2>/dev/null || echo "0")
      times+=($ms)
    else
      errors=$((errors + 1))
    fi
  done
  
  if [[ ${#times[@]} -gt 0 ]]; then
    sum=0
    min=${times[0]}
    max=${times[0]}
    
    for t in "${times[@]}"; do
      sum=$(echo "$sum + $t" | bc)
      [[ $(echo "$t < $min" | bc -l) ]] && min=$t
      [[ $(echo "$t > $max" | bc -l) ]] && max=$t
    done
    
    avg=$(echo "scale=0; $sum / ${#times[@]}" | bc)
    
    sorted=($(for a in "${times[@]}"; do echo "$a"; done | sort -n))
    count=${#times[@]}
    p50_idx=$(( (count - 1) / 2 ))
    p95_idx=$(( (count * 95 / 100) ))
    [[ $p95_idx -ge $count ]] && p95_idx=$((count - 1))
    
    p50=${sorted[$p50_idx]}
    p95=${sorted[$p95_idx]}
    
    printf "  %-15s Avg: %5sms | P50: %sms | P95: %sms | Errors: %d\n" "$name:" "$avg" "$p50" "$p95" "$errors"
    
    echo "$name,http://localhost:$port/health,$avg,$p50,$p95,$min,$max,$errors" >> "$SCRIPT_DIR/latency_results.csv"
  fi
done

echo ""
log "Results saved to latency_results.csv"