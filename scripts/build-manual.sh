#!/usr/bin/env bash
#
# Membangun PDF manual pengguna dari markdown + screenshot.
#
# Generator PDF yang dipakai tidak bisa memuat berkas gambar dari disk, jadi
# tiap penanda {{img:nama}} pada markdown diganti dengan data URI base64 dari
# docs/manual-img/nama.png. Gambar dikecilkan dulu agar ukuran PDF wajar.
#
#   ./scripts/build-manual.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC="$ROOT/docs/manual-pengguna-ibatiks.md"
IMG="$ROOT/docs/manual-img"
OUT="$ROOT/docs/Manual-Pengguna-Ibatiks.pdf"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

PDF_BIN="${MAKE_PDF_BIN:-$HOME/.claude/skills/gstack/make-pdf/dist/pdf}"
[[ -x "$PDF_BIN" ]] || { echo "make-pdf tidak ditemukan di $PDF_BIN"; exit 1; }
command -v sips >/dev/null || { echo "butuh sips (bawaan macOS)"; exit 1; }

echo "Mengecilkan screenshot…"
mkdir -p "$WORK/img"
count=0
for f in "$IMG"/*.png; do
  [[ -e "$f" ]] || continue
  sips -Z 1000 "$f" --out "$WORK/img/$(basename "$f")" >/dev/null 2>&1
  count=$((count + 1))
done
echo "  $count gambar diproses"

echo "Menyematkan gambar ke markdown…"
python3 - "$SRC" "$WORK/img" "$WORK/manual.md" <<'PY'
import base64, pathlib, re, sys

src, imgdir, dst = (pathlib.Path(a) for a in sys.argv[1:4])
text = src.read_text(encoding="utf-8")
missing = []

def embed(match):
    name = match.group(1).strip()
    path = imgdir / f"{name}.png"
    if not path.exists():
        missing.append(name)
        return f"*(gambar {name} tidak ditemukan)*"
    data = base64.b64encode(path.read_bytes()).decode("ascii")
    # Ditulis sebagai tag HTML dengan lebar eksplisit, bukan sintaks gambar
    # markdown biasa: tanpa itu sebagian gambar dirender melebihi lebar kolom
    # lalu terpotong di tepi halaman PDF.
    return (
        f'<img src="data:image/png;base64,{data}" alt="{name}" '
        f'style="width:100%;height:auto;display:block;" />'
    )

out = re.sub(r"\{\{img:([^}]+)\}\}", embed, text)
dst.write_text(out, encoding="utf-8")

used = len(re.findall(r"\{\{img:[^}]+\}\}", text))
print(f"  {used - len(missing)} gambar tersemat")
if missing:
    print("  TIDAK DITEMUKAN: " + ", ".join(sorted(set(missing))))
    sys.exit(1)
PY

echo "Membuat PDF…"
"$PDF_BIN" generate "$WORK/manual.md" "$OUT" --cover --toc --no-confidential --quiet 2>/dev/null \
  || "$PDF_BIN" generate "$WORK/manual.md" "$OUT" --cover --toc --no-confidential >/dev/null

printf '\n✓ %s\n' "$OUT"
printf '  %s halaman, %s\n' \
  "$(pdfinfo "$OUT" 2>/dev/null | awk '/^Pages/{print $2}')" \
  "$(du -h "$OUT" | awk '{print $1}')"
