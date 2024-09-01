build-darwin:
	rustup target add aarch64-apple-darwin
	cargo build --target aarch64-apple-darwin --release
	mkdir -p ./libs
	cp ./target/aarch64-apple-darwin/release/libchia_capi.a ./libs/libchia_capi_aarch64-apple-darwin20.a

build-musl:
	rustup target add x86_64-unknown-linux-musl
	cargo build --target x86_64-unknown-linux-musl --release
	mkdir -p ./libs
	cp ./target/x86_64-unknown-linux-musl/release/libchia_capi.a ./libs/libchia_capi_x86_64-unknown-linux-musl.a

build-linux:
	rustup target add x86_64-unknown-linux-gnu
	cargo build --target x86_64-unknown-linux-gnu --release
	mkdir -p ./libs
	cp ./target/x86_64-unknown-linux-gnu/release/libchia_capi.a ./libs/libchia_capi_x86_64-unknown-linux-gnu.a

clean:
	rm -rf libs
	rm -rf `find -name "*.o"`
	cargo clean