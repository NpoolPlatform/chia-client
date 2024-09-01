// use chia_rs::clvmr::run_program::run_program;

// #[no_mangle]
// #[repr(C)]
// pub struct ChiaResult {
//     cost: u64,
//     node: *mut c_void,
//     error: i32, // 0 for success, non-zero for error codes
// }

// #[no_mangle]
// pub extern "C" fn run_chia_program_demo(
//     program_ptr: *const u8,
//     program_len: usize,
//     args_ptr: *const u8,
//     args_len: usize,
//     max_cost: u64,
//     flags: u32,
// ) -> ChiaResult {
//     let mut allocator = make_allocator(flags);
//     let max_cost = Cost::from(max_cost);
    
//     // 转换输入的字节切片
//     let program = unsafe { std::slice::from_raw_parts(program_ptr, program_len) };
//     let args = unsafe { std::slice::from_raw_parts(args_ptr, args_len) };

//     let deserialize = if (flags & ALLOW_BACKREFS) != 0 {
//         node_from_bytes_backrefs
//     } else {
//         node_from_bytes
//     };

//     // 反序列化程序和参数
//     let program_node = match deserialize(&mut allocator, program) {
//         Ok(node) => node,
//         Err(_) => return ChiaResult { cost: 0, node: ptr::null_mut(), error: 1 },
//     };

//     let args_node = match deserialize(&mut allocator, args) {
//         Ok(node) => node,
//         Err(_) => return ChiaResult { cost: 0, node: ptr::null_mut(), error: 1 },
//     };

//     let dialect = ChiaDialect::new(flags);
    
//     // 运行程序
//     let reduction_result = run_program(&mut allocator, &dialect, program_node, args_node, max_cost);

//     match reduction_result {
//         Ok(reduction) => {
//             let val = LazyNode::new(Rc::new(allocator), reduction.1);
//             let node_ptr = Rc::into_raw(Rc::new(val)) as *mut c_void;
//             ChiaResult {
//                 cost: reduction.0.get(),
//                 node: node_ptr,
//                 error: 0,
//             }
//         },
//         Err(e) => {
//             ChiaResult { cost: 0, node: ptr::null_mut(), error: 2 }
//         }
//     }
// }

#[no_mangle]
pub extern "C" fn add(left: i32, right: i32) -> i32 {
    left + right
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn it_works() {
        let result = add(2, 2);
        assert_eq!(result, 4);
    }
}
