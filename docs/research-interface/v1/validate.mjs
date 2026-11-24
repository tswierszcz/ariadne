#!/usr/bin/env node
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';

// Documentation conformance check. Requires AJV 8 and ajv-formats.
const require = createRequire(import.meta.url);
const Ajv = require('ajv/dist/2020.js');
const addFormats = require('ajv-formats');
const root = path.dirname(fileURLToPath(import.meta.url));
const schema = JSON.parse(fs.readFileSync(path.join(root, 'contract.schema.json'), 'utf8'));
const ajv = new Ajv({ allErrors: true, strict: true });
addFormats(ajv);
ajv.addSchema(schema);

function readJson(relativePath) {
  return JSON.parse(fs.readFileSync(path.join(root, relativePath), 'utf8'));
}

function sha256(bytes) {
  return `sha256:${crypto.createHash('sha256').update(bytes).digest('hex')}`;
}

function checkSchema(definition, value) {
  const validate = ajv.getSchema(`${schema.$id}#/$defs/${definition}`);
  assert(validate, `Unknown schema definition ${definition}`);
  assert(validate(value), `${definition}: ${ajv.errorsText(validate.errors)}`);
}

function safePath(directory, relativePath) {
  assert(!path.isAbsolute(relativePath), 'Absolute bundle member');
  const resolved = path.resolve(directory, relativePath);
  assert(resolved.startsWith(`${path.resolve(directory)}${path.sep}`), 'Bundle path escapes directory');
  let current = path.resolve(directory);
  for (const part of path.relative(directory, resolved).split(path.sep)) {
    current = path.join(current, part);
    assert(!fs.lstatSync(current).isSymbolicLink(), 'Symlink bundle member');
  }
  return resolved;
}

function indexRecords(records) {
  const result = { nodes: new Map(), edges: new Map(), roots: new Map(), gaps: new Map(),
    components: new Map(), classes: new Map(), calls: new Map() };
  const fields = {
    node: ['nodes', 'node_id'], edge: ['edges', 'edge_id'], root: ['roots', 'root_id'],
    gap: ['gaps', 'gap_id'], component: ['components', 'instance_id'],
    class: ['classes', 'class_id'], call: ['calls', 'call_id'],
  };
  for (const record of records) {
    checkSchema('Record', record);
    const [collection, idField] = fields[record.record_type];
    assert(!result[collection].has(record[idField]), `Duplicate ${idField}`);
    result[collection].set(record[idField], record);
  }
  return result;
}

function checkSite(site, index) {
  if (site !== null) assert(index.nodes.has(site.caller_node_id), 'Unknown site caller');
}

function checkGraph(records) {
  const index = indexRecords(records);
  for (const record of index.classes.values()) {
    assert(index.components.has(record.instance_id), 'Unknown class instance');
    const selected = index.classes.get(record.selected_class_id);
    assert(selected?.selected && selected.class === record.class, 'Invalid selected class');
    if (record.selected) assert.equal(record.class_id, record.selected_class_id);
  }
  for (const node of index.nodes.values()) {
    assert(index.nodes.has(node.canonical_node_id), 'Unknown canonical node');
    if (node.node_kind === 'synthetic_event') {
      assert.equal(node.method, null);
      assert.equal(node.body, 'synthetic');
    } else {
      assert(node.method, 'Method node requires method identity');
      if (node.body !== 'external') {
        assert(index.components.has(node.method.instance_id), 'Unknown method instance');
        const owner = [...index.classes.values()].find(record => record.selected
          && record.instance_id === node.method.instance_id
          && record.class === node.method.class
          && record.class_sha256 === node.method.class_sha256);
        assert(owner, 'Method has no selected owning class');
        assert.notEqual(node.access_flags, null);
      }
    }
    if (node.node_kind === 'receiver_specialization') {
      assert(node.receiver_type, 'Specialization requires receiver');
      assert.notEqual(node.canonical_node_id, node.node_id);
    } else {
      assert.equal(node.canonical_node_id, node.node_id);
      assert.equal(node.receiver_type, null);
    }
    const source = node.source;
    if (source.line_start !== null && source.line_end !== null) {
      assert(source.line_end >= source.line_start, 'Reversed source range');
    }
  }
  for (const edge of index.edges.values()) {
    assert(index.nodes.has(edge.from) && index.nodes.has(edge.to), 'Unknown edge endpoint');
    for (const site of edge.sites) {
      checkSite(site, index);
      assert.equal(site.caller_node_id, edge.from);
    }
  }
  for (const entry of index.roots.values()) assert(index.nodes.has(entry.node_id), 'Unknown root');
  for (const gap of index.gaps.values()) {
    checkSite(gap.site, index);
    for (const instance of gap.candidate_instance_ids ?? []) {
      assert(index.components.has(instance), 'Unknown gap instance');
    }
    for (const method of gap.candidate_method_ids ?? []) {
      assert([...index.nodes.values()].some(node => node.method?.method_id === method), 'Unknown gap method');
    }
  }
  for (const call of index.calls.values()) {
    checkSite(call.site, index);
    for (const id of call.target_node_ids) assert(index.nodes.has(id), 'Unknown call target');
  }
  return index;
}

function checkBundle() {
  const directory = path.join(root, 'fixtures/bundle');
  const manifest = readJson('fixtures/bundle/manifest.json');
  checkSchema('Manifest', manifest);
  const identityBytes = fs.readFileSync(safePath(directory, manifest.identity_path));
  assert.equal(sha256(identityBytes), manifest.analysis_id, 'Identity hash mismatch');
  const identity = JSON.parse(identityBytes);
  checkSchema('Identity', identity);
  assert.equal(identity.contract_revision, manifest.contract_revision);
  const inputIds = identity.inputs.map(input => input.id);
  assert.equal(new Set(inputIds).size, inputIds.length, 'Duplicate input id');
  assert(identity.classpath_order.every(id => inputIds.includes(id)), 'Unknown classpath input');
  const datasetPaths = new Set();
  const records = [];
  for (const dataset of identity.datasets) {
    assert(!datasetPaths.has(dataset.path), 'Duplicate dataset');
    datasetPaths.add(dataset.path);
    const bytes = fs.readFileSync(safePath(directory, dataset.path));
    assert.equal(bytes.length, dataset.size_bytes);
    assert.equal(sha256(bytes), dataset.sha256);
    const text = bytes.toString('utf8');
    assert(text.endsWith('\n'), 'JSONL must end with LF');
    const rows = text.slice(0, -1).split('\n').map(line => JSON.parse(line));
    assert.equal(rows.length, dataset.record_count);
    records.push(...rows);
  }
  const index = checkGraph(records);
  return { manifest, identity, records, index };
}

const bundle = checkBundle();
const indexedRoots = new Set([...bundle.index.roots.values()].map(record => record.node_id));

function checkRequest(request) {
  assert.equal(request.analysis_id, bundle.manifest.analysis_id, 'Request analysis mismatch');
  if (request.operation === 'reachability') {
    const ids = request.targets.map(target => target.id);
    assert.equal(new Set(ids).size, ids.length, 'Duplicate target id');
    if (request.roots.selection === 'subset') {
      assert(request.roots.node_ids.every(id => indexedRoots.has(id)), 'Root outside index');
    }
  }
  const accepted = {
    entries: ['instance_id', 'class_prefix', 'method_name', 'descriptor', 'node_ids'],
    methods: ['instance_id', 'class_prefix', 'method_name', 'descriptor', 'node_ids', 'reachable_only'],
    calls: ['instance_id', 'class_prefix', 'method_name', 'descriptor', 'node_ids', 'declared_class_prefix', 'reachable_only'],
  };
  for (const field of Object.keys(request.filters ?? {})) {
    assert(accepted[request.operation]?.includes(field), 'Filter unsupported for operation');
  }
}

function checkResponse(response, request) {
  if (response.status === 'error') return;
  assert.equal(response.analysis_id, bundle.manifest.analysis_id, 'Response analysis mismatch');
  assert.deepEqual(response.provenance, bundle.identity.provenance);
  if (request) {
    assert.equal(response.request_id, request.request_id);
    assert.equal(response.operation, request.operation);
    assert.equal(response.analysis_id, request.analysis_id);
  }
  if (response.operation !== 'reachability') {
    assert.equal(response.page.returned_count, response.records.length);
    const types = { entries: 'root', methods: 'node', calls: 'call', callers: 'edge', callees: 'edge' };
    assert(response.records.every(record => record.record_type === types[response.operation]));
    for (const record of response.records) assertOriginalRecord(record);
    assert(response.records.length <= response.effective_limits.max_records);
    if (!response.page.truncated) assert.equal(response.page.next_cursor, null);
    return;
  }
  const evidence = indexRecords(response.evidence);
  for (const record of response.evidence) assertOriginalRecord(record);
  const expectedIds = request?.targets.map(target => target.id);
  if (expectedIds) assert.deepEqual(response.results.map(result => result.target_id), expectedIds);
  let totalWork = 0;
  for (const result of response.results) {
    totalWork += result.search.work_units;
    for (const id of result.matched_node_ids) assert(bundle.index.nodes.has(id), 'Unknown matched node');
    for (const id of result.model.relevant_gap_ids) assert(evidence.gaps.has(id), 'Missing gap evidence');
    if (result.resolution !== 'matched') {
      assert.equal(result.search.state, 'not_evaluated');
      assert.equal(result.entries.length, 0);
      assert.equal(result.model.state, 'undetermined');
    }
    if (result.resolution === 'missing' || result.resolution === 'shadowed') {
      assert.equal(result.matched_node_ids.length, 0);
    }
    if (result.resolution === 'matched') assert(result.matched_node_ids.length > 0);
    if (request) {
      assert.equal(result.root_basis, request.roots.selection === 'indexed'
        ? 'indexed_graph' : 'subset_on_indexed_graph');
    }
    if (result.search.state === 'found') {
      assert.equal(result.resolution, 'matched');
      assert(result.entries.length > 0, 'Found requires path evidence');
    } else {
      assert.equal(result.entries.length, 0);
    }
    if (result.search.state === 'not_found') {
      assert.equal(result.resolution, 'matched');
      assert.equal(result.search.work_complete, true);
      assert.equal(result.total_entry_count, 0);
    }
    if (result.search.state === 'incomplete') {
      assert.equal(result.search.work_complete, false);
      assert.equal(result.total_entry_count, null);
      assert(result.search.reason);
    }
    if (result.model.state === 'complete_under_declared_model') {
      assert.equal(result.model.relevant_gap_ids.length, 0);
      assert.equal(result.model.scope_uncertain, false);
    }
    assert(result.entries.length <= response.effective_limits.max_entries);
    if (result.total_entry_count !== null) assert(result.total_entry_count >= result.entries.length);
    for (const entry of result.entries) {
      const proof = entry.path;
      assert(indexedRoots.has(entry.root_node_id));
      if (request?.roots.selection === 'subset') assert(request.roots.node_ids.includes(entry.root_node_id));
      assert.equal(proof.node_ids[0], entry.root_node_id);
      assert.equal(proof.edge_ids.length, proof.node_ids.length - 1);
      assert(proof.node_ids.length <= response.effective_limits.max_path_nodes);
      if (!proof.truncated) {
        assert(result.matched_node_ids.includes(proof.node_ids.at(-1)), 'Path ends outside target');
        assert.equal(proof.full_node_count, proof.node_ids.length);
      }
      for (const id of proof.node_ids) assert(evidence.nodes.has(id), 'Missing node evidence');
      proof.edge_ids.forEach((id, position) => {
        const edge = evidence.edges.get(id);
        assert(edge, 'Missing edge evidence');
        assert.equal(edge.from, proof.node_ids[position]);
        assert.equal(edge.to, proof.node_ids[position + 1]);
      });
    }
  }
  assert(totalWork <= response.effective_limits.max_work_units, 'Work budget exceeded');
}

function assertOriginalRecord(record) {
  const idFields = { component: 'instance_id', class: 'class_id', node: 'node_id',
    edge: 'edge_id', root: 'root_id', gap: 'gap_id', call: 'call_id' };
  const field = idFields[record.record_type];
  const original = bundle.records.find(candidate => candidate.record_type === record.record_type
    && candidate[field] === record[field]);
  assert(original, 'Response record is absent from the bundle');
  assert.deepEqual(record, original, 'Response record changed the frozen evidence');
}

const suite = readJson('cases.json');
let checked = 0;
for (const test of suite.cases) {
  const value = readJson(test.path);
  let schemaValid = true;
  try { checkSchema(test.definition, value); } catch { schemaValid = false; }
  assert.equal(schemaValid, test.schema_valid, `${test.id}: unexpected schema verdict`);
  if (schemaValid) {
    let semanticValid = true;
    try {
      if (test.definition === 'Request') checkRequest(value);
      if (test.definition === 'Response') checkResponse(value, test.request ? readJson(test.request) : null);
    } catch { semanticValid = false; }
    assert.equal(semanticValid, test.semantic_valid, `${test.id}: unexpected semantic verdict`);
  }
  checked++;
}
console.log(`Validated schema, bundle hashes, graph references and ${checked} positive/negative contract cases.`);
