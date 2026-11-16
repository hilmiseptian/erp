export function formatDateLabel(value, locale = "en-US") {
  if (typeof value !== "string" || !value.trim()) return "—";

  const datePart = value.slice(0, 10);
  if (!/^\d{4}-\d{2}-\d{2}$/.test(datePart)) return "—";

  const [year, month, day] = datePart.split("-").map(Number);
  const date = new Date(Date.UTC(year, month - 1, day));
  if (date.toISOString().slice(0, 10) !== datePart) return "—";

  return new Intl.DateTimeFormat(locale, { dateStyle: "medium", timeZone: "UTC" }).format(date);
}

export function formatCurrency(value, locale = "id-ID", currency = "IDR") {
  return new Intl.NumberFormat(locale, { style: "currency", currency }).format(Number(value || 0));
}
