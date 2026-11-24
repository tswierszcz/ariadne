#!/usr/bin/env node
// RESEARCH_INTERFACE: checks a real research bundle and response files
// against the shipped JSON Schema (docs/research-interface/v1): the
// manifest, the identity, every JSONL record, and each response file
// given after the bundle. Shape only; the Go reader (internal/research)
// enforces hashes and references. Needs Node.js with AJV 8 and
// ajv-formats, as docs/research-interface/v1/validate.mjs does.
//
// Usage: node scripts/research-check.mjs BUNDLE_DIR [RESPONSE.json ...]
import fs from 'node:fs';
import path from 'node:path';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';

const require = createRequire(import.meta.url);
const Ajv = require('ajv/dist/2020.js');
const addFormats = require('ajv-formats');
const here = path.dirname(fileURLToPath(import.meta.url));
const schema = JSON.parse(fs.readFileSync(path.join(here, '../docs/research-interface/v1/contract.schema.json'), 'utf8'));
const ajv = new Ajv({ allErrors: true, strict: true });
addFormats(ajv);
ajv.addSchema(schema);
const check = (definition, doc, where) => {
  const validate = ajv.getSchema(`${schema.$id}#/$defs/${definition}`);
  if (!validate(doc)) {
    console.error(`${where}: not a valid ${definition}:`, JSON.stringify(validate.errors.slice(0, 3)));
    process.exitCode = 1;
    return false;
  }
  return true;
};
const [bundle, ...responses] = process.argv.slice(2);
if (!bundle) { console.error('usage: research-check.mjs BUNDLE_DIR [RESPONSE.json ...]'); process.exit(2); }
const manifest = JSON.parse(fs.readFileSync(path.join(bundle, 'manifest.json'), 'utf8'));
check('Manifest', manifest, 'manifest.json');
const identity = JSON.parse(fs.readFileSync(path.join(bundle, manifest.identity_path), 'utf8'));
check('Identity', identity, manifest.identity_path);
let records = 0, bad = 0;
for (const ds of identity.datasets) {
  const lines = fs.readFileSync(path.join(bundle, ds.path), 'utf8').split('\n').filter(l => l);
  for (const [i, line] of lines.entries()) {
    records++;
    if (!check('Record', JSON.parse(line), `${ds.path}:${i + 1}`) && ++bad > 20) process.exit(1);
  }
}
for (const r of responses) check('Response', JSON.parse(fs.readFileSync(r, 'utf8')), r);
if (!process.exitCode) console.log(`schema-valid: manifest, identity, ${records} records, ${responses.length} responses`);
