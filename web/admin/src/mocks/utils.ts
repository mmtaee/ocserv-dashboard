export function cloneMock<T>(value: T): T {
  return structuredClone(value);
}
