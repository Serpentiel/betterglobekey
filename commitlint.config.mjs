export default {
  extends: ["@commitlint/config-conventional"],
  ignores: [
    (msg) => /Signed-off-by: dependabot\[bot]/m.test(msg),
    (msg) => /^\w+\(deps\): bump .+ from .+ to .+/.test(msg),
  ],
};
