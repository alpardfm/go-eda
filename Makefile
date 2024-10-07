.PHONY: post-products

URL=http://localhost:8080/products
DATA='{"name": "Product A", "price": 100}'

post-products:
	@start=$$(date +%s); \
	for i in $$(seq 1 1000); do \
		echo "Sending request $$i..."; \
		curl -X POST $(URL) -H "Content-Type: application/json" -d $(DATA); \
		echo ""; \
	done; \
	end=$$(date +%s); \
	duration=$$((end - start)); \
	echo "Total time taken: $$duration seconds"
