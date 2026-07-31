import { describe, expect, it } from "vitest";
import { buildSpec, emptyForm, parsePort, parsePorts, parseLabels, specToForm } from "@/features/create/specForm";

describe("parsePort", () => {
  it("parses host:container/proto and shorter forms", () => {
    expect(parsePort("8080:80/tcp")).toEqual({ host: "8080", container: 80, proto: "tcp" });
    expect(parsePort("8080:80")).toEqual({ host: "8080", container: 80, proto: "tcp" });
    expect(parsePort("53/udp")).toEqual({ host: "", container: 53, proto: "udp" });
    expect(parsePort("80")).toEqual({ host: "", container: 80, proto: "tcp" });
    expect(parsePort("nonsense")).toBeNull();
  });
});

describe("buildSpec", () => {
  it("parses list fields from text", () => {
    const spec = buildSpec({
      ...emptyForm,
      image: "nginx",
      name: "web",
      command: "nginx -g daemon",
      env: "TZ=UTC\nDEBUG=1",
      ports: "8080:80/tcp\n443",
      labels: "team=web\nenv=prod",
      restartPolicy: "always",
    });
    expect(spec.image).toBe("nginx");
    expect(spec.command).toEqual(["nginx", "-g", "daemon"]);
    expect(spec.env).toEqual(["TZ=UTC", "DEBUG=1"]);
    expect(spec.ports).toHaveLength(2);
    expect(spec.labels).toEqual({ team: "web", env: "prod" });
    expect(spec.restartPolicy).toBe("always");
  });

  it("empty command becomes null", () => {
    expect(buildSpec(emptyForm).command).toBeNull();
  });
});

describe("parseLabels", () => {
  it("ignores lines without =", () => {
    expect(parseLabels("a=1\ngarbage\nb=2")).toEqual({ a: "1", b: "2" });
  });
});

describe("specToForm round-trip", () => {
  it("rebuilds equivalent form text", () => {
    const original = buildSpec({ ...emptyForm, image: "redis", ports: "6379:6379/tcp", env: "X=1" });
    const form = specToForm(original);
    const rebuilt = buildSpec(form);
    expect(rebuilt.ports).toEqual(original.ports);
    expect(rebuilt.env).toEqual(original.env);
    expect(rebuilt.image).toBe("redis");
  });
});

describe("parsePorts", () => {
  it("skips invalid lines", () => {
    expect(parsePorts("80\nbad\n90/udp")).toHaveLength(2);
  });
});
