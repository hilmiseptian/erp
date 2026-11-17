export function formatDateLabel(value, locale = "en-US") {
  if (typeof value !== "string" || !value.trim()) return "—";

  const datePart = value.slice(0, 10);
  if (!/^\d{4}-\d{2}-\d{2}$/.test(datePart)) return "—";

  const [year, month, day] = datePart.split("-").map(Number);
  const date = new Date(Date.UTC(year, month - 1, day));
  if (date.toISOString().slice(0, 10) !== datePart) return "—";

  return new Intl.DateTimeFormat(locale, { dateStyle: "medium", timeZone: "UTC" }).format(date);
}

export function toDateInputValue(value) {
  if (typeof value !== "string" || !value.trim()) return "";

  const datePart = value.slice(0, 10);
  if (!/^\d{4}-\d{2}-\d{2}$/.test(datePart)) return "";

  const [year, month, day] = datePart.split("-").map(Number);
  const date = new Date(Date.UTC(year, month - 1, day));
  return date.toISOString().slice(0, 10) === datePart ? datePart : "";
}

export function formatCurrency(value, locale = "id-ID", currency = "IDR", maximumFractionDigits) {
  const options = { style: "currency", currency };
  if (maximumFractionDigits !== undefined) options.maximumFractionDigits = maximumFractionDigits;
  return new Intl.NumberFormat(locale, options).format(Number(value || 0));
}
