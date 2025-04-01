import { user } from '../stores/user';
import { api } from './api';
import { invoice } from '@telegram-apps/sdk-svelte';

export async function openTelegramInvoice(starsAmount: number) {
    try {
        console.log('Creating invoice for', starsAmount, 'Stars');
        const data = await api.createInvoice(starsAmount);
        console.log('Invoice data:', data);
        
        if (!data.url) {
            throw new Error('No invoice URL in response');
        }

        if (!invoice.isSupported()) {
            console.error('Invoices are not supported in this version of Telegram');
            return;
        }

        const status = await invoice.open(data.url, 'url');
        console.log('Payment status:', status);
        
        switch (status) {
            case 'paid':
                console.log('Оплата прошла успешно');
                window.location.reload();
                break;
            case 'failed':
                console.log('Ошибка оплаты');
                break;
            case 'cancelled':
                console.log('Оплата отменена');
                window.location.reload();
                break;
            case 'pending':
                console.log('Оплата в процессе');
                break;
            default:
                console.log('Неизвестный статус оплаты:', status);
        }
    } catch (error) {
        console.error('Ошибка при создании инвойса:', error);
    }
}