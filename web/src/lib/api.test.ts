import { describe, expect, it } from "vitest";
import { humanSize, primaryName, primaryTag, shortId } from "@/lib/api";

describe("formatting helpers", () => {
  it("shortId strips sha256 prefix and truncates", () => {
    expect(shortId("sha256:0123456789abcdef")).toBe("0123456789ab");
    expect(shortId("abc")).toBe("abc");
  });

  it("humanSize uses SI units", () => {
    expect(humanSize(999)).toBe("999 B");
    expect(humanSize(1_500_000)).toBe("1.5 MB");
    expect(humanSize(2_000_000_000)).toBe("2.0 GB");
  });

  it("primaryName/primaryTag fall back to <none>", () => {
    expect(primaryName(["web", "web2"])).toBe("web");
    expect(primaryName(null)).toBe("<none>");
    expect(primaryTag([])).toBe("<none>");
    expect(primaryTag(["nginx:latest"])).toBe("nginx:latest");
  });
});
