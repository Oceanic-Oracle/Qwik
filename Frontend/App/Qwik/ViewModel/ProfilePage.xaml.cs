using Qwik.Model;
using Qwik.Model.Dto;

namespace Qwik;

public partial class ProfilePage : ContentPage
{
    string Login = string.Empty;
    string Email = string.Empty;
    string CreatedAt = string.Empty;
    public ProfilePage()
	{
		InitializeComponent();
    }

    protected override async void OnAppearing()
	{
        if (string.IsNullOrEmpty(Config.JWT))
        {
            await Shell.Current.Navigation.PushAsync(new AuthPage());
            return;
        } else
        {
            try
            {
                var header = new Dictionary<string, string>();
                header.Add("Authorization", $"Bearer {Config.JWT}");

                var body = await Api.SendGetRequest<GetProfileRes>(header, Api.MyProfileEndpoint);
                CreatedAt = body.CreatedAt;
            }
            catch (Exception ex)
            {
                await DisplayAlert("Ошибка", $"{ex.Message}", "OK");

                Config.JWT = ex.Message;
                await Shell.Current.Navigation.PushAsync(new AuthPage());
            }
        }
    }

    private void OnSearchClicked(object sender, EventArgs e)
    {
        
    }

    private async void OpenProfileClicked(object sender, EventArgs e)
    {
        bool action = await DisplayAlert(
            "Профиль",
            $"Логин: {Login}\nПочта: {Email}\nДата регистрации: {CreatedAt}",
            "Выйти",
            "Отмена");

        if (action)
        {
            Config.JWT = "";
            await Shell.Current.Navigation.PushAsync(new AuthPage());
        }
    }
}