import matplotlib.pyplot as plt
# Actual Test Data
threads = [1, 2, 4, 8]
times = [32.83, 15.95, 9.56, 6.31]

# Automatically calculate speedup (Speedup = Base Time / Current Time)
base_time = times[0]
speedups = [base_time / t for t in times]
# Plot Settings
plt.figure(figsize=(10, 6))

# Plot ideal linear speedup reference line (y=x)
plt.plot(threads, threads, 'k--', label='Ideal Linear Speedup', alpha=0.5)

# Plot your actual speedup
plt.plot(threads, speedups, marker='o', linewidth=2, markersize=8, color='#007acc', label='Actual Speedup')

# Annotate specific values on each point
for x, y in zip(threads, speedups):
    plt.annotate(f'{y:.2f}x', 
                 (x, y), 
                 textcoords="offset points", 
                 xytext=(0,10), 
                 ha='center',
                 fontsize=10,
                 fontweight='bold')

# Set title and labels
plt.title('Wa-Tor Simulation Parallel Speedup', fontsize=14)
plt.xlabel('Number of Threads (Workers)', fontsize=12)
plt.ylabel('Speedup Factor', fontsize=12)

# Set axis ticks
plt.xticks(threads)
plt.grid(True, linestyle='--', alpha=0.7)
plt.legend()

# Save the Image
output_filename = 'results_graph.png'
plt.savefig(output_filename, dpi=300)
print(f"Success! Graph saved as {output_filename}")