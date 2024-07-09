export function allocate(len: i32): usize {
  // create a new AssemblyScript byte array
  let buf = new Array<u8>(len);
  let buf_ptr = heap.alloc(len * 2); // serve enough space
  // create a pointer to the byte array and
  // return it
  store<Array<u8>>(buf_ptr, buf);
  return buf_ptr;
}
