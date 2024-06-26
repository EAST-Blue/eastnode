export function allocate(len: i32): usize {
  // create a new AssemblyScript byte array
  let buf = new Array<u8>(len);
  let buf_ptr = memory.data(8);
  // create a pointer to the byte array and
  // return it
  store<Array<u8>>(buf_ptr, buf);
  return buf_ptr;
}

export function toString(ptr: i64): string {
  // get length
  let len = load<u32>(usize(ptr - 4));
  return String.UTF16.decodeUnsafe(<usize>ptr, <usize>len);
}