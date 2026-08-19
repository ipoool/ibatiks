/**
 * Pesan validasi bawaan browser, dalam Bahasa Indonesia.
 *
 * Browser menyusun pesannya sendiri untuk atribut seperti `required`, `min`,
 * dan `type="email"` — dan bahasanya mengikuti bahasa antarmuka browser, bukan
 * `lang` halaman. Di aplikasi yang seluruhnya berbahasa Indonesia, tim toko
 * mendapat gelembung berbunyi "Please fill out this field." di tengah formulir
 * yang semua labelnya Indonesia.
 *
 * Pesan di sini menggantikannya, sekaligus menyebutkan nama kolomnya kalau ada
 * labelnya — sesuatu yang tidak dilakukan pesan bawaan browser.
 */

type Kolom = HTMLInputElement | HTMLTextAreaElement;

/** Nama kolom dari labelnya, tanpa tanda bintang penanda wajib. */
function namaKolom(el: Kolom): string | null {
  const teks = el.labels?.[0]?.textContent?.replace(/\*/g, "").trim();
  return teks ? teks : null;
}

function angka(nilai: string | null): string {
  const n = Number(nilai);
  return Number.isFinite(n) ? n.toLocaleString("id-ID") : (nilai ?? "");
}

/**
 * Batas `min`/`max` seperti dibacakan orang.
 *
 * Kolom tanggal menyimpan batasnya dalam bentuk ISO (2026-09-20). Menyalinnya
 * apa adanya ke pesan galat memaksa admin membaca format yang tidak dipakai di
 * mana pun lagi dalam aplikasi ini.
 */
function batas(el: Kolom, atribut: "min" | "max"): string {
  const nilai = el.getAttribute(atribut);
  if (!nilai) return "";
  if (el.type === "date") {
    const t = new Date(`${nilai}T00:00:00`);
    if (!Number.isNaN(t.getTime())) {
      return t.toLocaleDateString("id-ID", { day: "numeric", month: "long", year: "numeric" });
    }
  }
  return angka(nilai);
}

/**
 * Pesan untuk keadaan kolom saat ini. Dipanggil setelah customValidity dibuang,
 * jadi yang terbaca murni hasil pemeriksaan browser.
 */
export function pesanValidasi(el: Kolom): string {
  const v = el.validity;
  const nama = namaKolom(el);
  const sebut = nama ? `${nama} ` : "";

  if (v.valueMissing) {
    return nama ? `${nama} wajib diisi.` : "Kolom ini wajib diisi.";
  }
  if (v.typeMismatch) {
    if (el.type === "email") return "Format email belum benar, contohnya nama@toko.id.";
    if (el.type === "url") return "Alamat tautan belum benar, awali dengan https://.";
    return `${sebut}belum sesuai format yang diminta.`;
  }
  if (v.patternMismatch) {
    return `${sebut}belum sesuai format yang diminta.`;
  }
  if (v.tooShort) {
    return `${sebut}minimal ${el.minLength} karakter, baru terisi ${el.value.length}.`;
  }
  if (v.tooLong) {
    return `${sebut}maksimal ${el.maxLength} karakter.`;
  }
  if (v.rangeUnderflow) {
    return `${sebut}minimal ${batas(el, "min")}.`;
  }
  if (v.rangeOverflow) {
    return `${sebut}maksimal ${batas(el, "max")}.`;
  }
  if (v.stepMismatch) {
    return `${sebut}tidak boleh berkoma sebanyak itu.`;
  }
  if (v.badInput) {
    return `${sebut}belum bisa dibaca, periksa lagi ketikannya.`;
  }
  return `${sebut}belum benar.`;
}

/**
 * Memasang pesan Indonesia pada gelembung validasi kolom.
 *
 * customValidity dibuang lebih dulu supaya penanda bawaan browser (valueMissing
 * dan kawan-kawannya) terbaca apa adanya, bukan tertutup pesan dari percobaan
 * kirim sebelumnya.
 */
export function pasangPesanValidasi(el: Kolom): void {
  el.setCustomValidity("");
  el.setCustomValidity(pesanValidasi(el));
}

/**
 * Membuang pesan yang sudah dipasang.
 *
 * Wajib dipanggil begitu isinya berubah: selama customValidity masih terisi,
 * kolomnya dianggap tidak sah walaupun sudah dibetulkan, dan formulirnya tidak
 * akan pernah mau dikirim.
 */
export function bersihkanPesanValidasi(el: Kolom): void {
  if (el.validationMessage) el.setCustomValidity("");
}
