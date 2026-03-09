#!/usr/bin/env bash
# Fix duplicate migration versions by renumbering "split" pairs (e.g. 000017_foo.down + 000018_foo.up) to new versions.
set -e
cd "$(dirname "$0")/../migrations"
next=206
for v in $(ls *.sql 2>/dev/null | sed 's/_.*//' | sort -u); do
  [[ ! "$v" =~ ^[0-9]+$ ]] && continue
  files=($(ls ${v}_*.sql 2>/dev/null))
  [ ${#files[@]} -le 2 ] && continue
  # Find the migration name that appears only once in this version (the orphan)
  for f in "${files[@]}"; do
    name=$(basename "$f" .sql | sed "s/^${v}_//" | sed 's/\.up$//' | sed 's/\.down$//')
    up="${v}_${name}.up.sql"
    down="${v}_${name}.down.sql"
    if [ ! -f "$up" ] || [ ! -f "$down" ]; then
      # Orphan: find partner in v+1 or v-1
      vn=$(printf "%06d" $((10#$v + 1)))
      vp=$(printf "%06d" $((10#$v - 1)))
      if [ -f "${vn}_${name}.up.sql" ] || [ -f "${vn}_${name}.down.sql" ]; then
        partner_dir="$vn"
      elif [ -f "${vp}_${name}.up.sql" ] || [ -f "${vp}_${name}.down.sql" ]; then
        partner_dir="$vp"
      else
        continue
      fi
      new=$(printf "%06d" $next)
      if [ -f "${v}_${name}.down.sql" ]; then
        mv "${v}_${name}.down.sql" "${new}_${name}.down.sql"
        mv "${partner_dir}_${name}.up.sql" "${new}_${name}.up.sql"
      else
        mv "${v}_${name}.up.sql" "${new}_${name}.up.sql"
        mv "${partner_dir}_${name}.down.sql" "${new}_${name}.down.sql"
      fi
      echo "Renumbered ${name} -> ${new}"
      next=$((next + 1))
      break
    fi
  done
done
