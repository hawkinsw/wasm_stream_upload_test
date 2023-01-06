// Shamelessly stolen from https://dev.to/taterbase/running-a-go-program-in-deno-via-wasm-2l08
import * as _ from "./wasm_exec.js";
const go = new window.Go();
const f = await Deno.open("./web_client/client.wasm")
const buf = await Deno.readAll(f);
const inst = await WebAssembly.instantiate(buf, go.importObject);
go.run(inst.instance);
