<script lang="ts">
  import { onMount } from 'svelte';
  import { _ } from 'svelte-i18n';
  export let data: { date: string; count: number }[] = [];
  
  const months = ['jan', 'feb', 'mar', 'apr', 'may', 'jun', 'jul', 'aug', 'sep', 'oct', 'nov', 'dec'];
  const days = ['monday', 'wednesday', 'friday'];
  
  let scrollContainer: HTMLElement;
  
  onMount(() => {
    if (scrollContainer) {
      scrollContainer.scrollLeft = scrollContainer.scrollWidth;
    }
  });

  let tooltip = {
    text: '',
    visible: false,
    x: 0,
    y: 0
  };
  
  function getColorForCount(count: number): string {
    if (count === 0) return '#ebedf0';
    if (count <= 3) return '#9be9a8';
    if (count <= 6) return '#40c463';
    if (count <= 9) return '#30a14e';
    return '#216e39';
  }
  
  function showTooltip(event: MouseEvent, text: string) {
    const rect = (event.target as HTMLElement).getBoundingClientRect();
    tooltip = {
      text,
      visible: true,
      x: rect.left,
      y: rect.top - 30
    };
    
    setTimeout(() => {
      tooltip.visible = false;
    }, 2000);
  }
  
  let calendarData: { date: string; count: number }[] = [];
  let monthLabels: { text: string; column: number }[] = [];
  
  $: {
    console.log("Data changed:", data);
    const today = new Date();
    const oneYearAgo = new Date();
    oneYearAgo.setFullYear(today.getFullYear() - 1);
    
    // Устанавливаем дату на начало недели (понедельник)
    const startDate = new Date(oneYearAgo);
    startDate.setDate(startDate.getDate() - startDate.getDay() + 1);
    if (startDate.getDay() === 0) startDate.setDate(startDate.getDate() - 6);
    
    const calendar = [];
    let currentDate = new Date(startDate);
    let currentColumn = 0;
    let lastMonth = -1;
    monthLabels = [];
    
    while (currentDate <= today) {
      const dateStr = currentDate.toISOString().split('T')[0];
      const activity = data.find(d => d.date === dateStr);
      
      // Проверяем, начался ли новый месяц
      if (currentDate.getMonth() !== lastMonth) {
        monthLabels.push({
          text: months[currentDate.getMonth()],
          column: currentColumn
        });
        lastMonth = currentDate.getMonth();
      }
      
      calendar.push({
        date: dateStr,
        count: activity ? activity.count : 0
      });
      
      // Переходим к следующему дню и увеличиваем счетчик колонки каждые 7 дней
      currentDate.setDate(currentDate.getDate() + 1);
      if (calendar.length % 7 === 0) {
        currentColumn++;
      }
    }
    console.log("Generated calendar data:", calendar);
    calendarData = calendar;
  }
</script>

<div class="activity-heatmap">
  <div class="scroll-container" bind:this={scrollContainer}>
    <div class="heatmap-content">
      <div class="months">
        {#each monthLabels as { text, column }}
          <div class="month" style="grid-column: {column + 1}">
            {$_(`months.${text}`)}
          </div>
        {/each}
      </div>
      
      <div class="calendar">
        <div class="days">
          {#each days as day}
            <div class="day">{$_(`days.${day}`)}</div>
          {/each}
        </div>
        
        <div class="squares">
          {#each calendarData as day}
            <button 
              class="square" 
              style="background-color: {getColorForCount(day.count)}"
              on:click={(e) => showTooltip(e, `${day.date}: ${day.count} contributions`)}
              type="button"
              aria-label="{day.date}: {day.count} contributions"
            ></button>
          {/each}
        </div>
      </div>
    </div>
  </div>
  
  <!-- <div class="legend">
    <span>Less</span>
    <div class="square" style="background-color: #ebedf0"></div>
    <div class="square" style="background-color: #9be9a8"></div>
    <div class="square" style="background-color: #40c463"></div>
    <div class="square" style="background-color: #30a14e"></div>
    <div class="square" style="background-color: #216e39"></div>
    <span>More</span>
  </div> -->
</div>

{#if tooltip.visible}
  <div 
    class="tooltip"
    style="left: {tooltip.x}px; top: {tooltip.y}px"
  >
    {tooltip.text}
  </div>
{/if}

<style>
  .activity-heatmap {
    font-size: 12px;
    padding: 1rem;
  }
  
  .scroll-container {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
    margin: 0 -1rem;
    /* padding: 0 1rem; */
  }

  .heatmap-content {
    min-width: min-content;
  }
  
  .months {
    display: grid;
    grid-template-columns: repeat(53, 1fr);
    text-align: start;
    margin-bottom: 0.5rem;
    min-width: 700px;
    margin-left: 20px;
    gap: 2px;
  }
  
  .month {
    grid-row: 1;
    margin-left: 8px;
  }
  
  .calendar {
    display: flex;
    gap: 0.5rem;
  }
  
  .days {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    width: 20px;
  }
  
  .squares {
    display: grid;
    grid-template-columns: repeat(53, 1fr);
    grid-auto-flow: column;
    grid-template-rows: repeat(7, 1fr);
    gap: 2px;
    min-width: 700px;
  }
  
  .square {
    width: 10px;
    height: 10px;
    border-radius: 2px;
    padding: 0;
    border: none;
    cursor: pointer;
  }
  
  .legend {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    margin-top: 1rem;
    justify-content: flex-end;
  }
  
  .legend .square {
    width: 12px;
    height: 12px;
    cursor: default;
  }

  .tooltip {
    position: fixed;
    background: rgba(0, 0, 0, 0.8);
    color: white;
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 12px;
    pointer-events: none;
    z-index: 1000;
  }
</style> 